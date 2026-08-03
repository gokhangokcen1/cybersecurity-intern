package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

type Device struct {
	DeviceName   string   `json:"device_name"`
	CustomerName string   `json:"customer_name"`
	DealerName   string   `json:"dealer_name"`
	IPAddresses  []string `json:"ip_addresses"`
}
type Feed struct {
	Devices []Device `json:"devices"`
}
type Target struct{ IP, DeviceName, CustomerName, DealerName string }
type ScanJob struct {
	Target Target
	Port   int
}
type BatchJob struct {
	ScanJob
	done      *sync.WaitGroup
	confirm   bool
	retryJobs *[]ScanJob
	retryMu   *sync.Mutex
}
type ScanRequest struct {
	StartPort     int    `json:"startPort"`
	EndPort       int    `json:"endPort"`
	Ports         []int  `json:"ports"`
	TimeoutMS     int    `json:"timeoutMs"`
	Workers       int    `json:"workers"`
	PortsPerBatch int    `json:"portsPerBatch"`
	Mode          string `json:"mode"`
}
type OpenPort struct {
	Source       string    `json:"source"`
	IP           string    `json:"ip"`
	DeviceName   string    `json:"deviceName"`
	CustomerName string    `json:"customerName"`
	DealerName   string    `json:"dealerName"`
	Port         int       `json:"port"`
	FoundAt      time.Time `json:"foundAt"`
}
type ScanEvent struct {
	Type      string    `json:"type"`
	Port      int       `json:"port"`
	Completed int       `json:"completed"`
	Total     int       `json:"total"`
	Result    *OpenPort `json:"result,omitempty"`
	Message   string    `json:"message,omitempty"`
}
type ScanStatus struct {
	ScannerName    string     `json:"scannerName"`
	Running        bool       `json:"running"`
	CurrentPort    int        `json:"currentPort"`
	Completed      int        `json:"completed"`
	Total          int        `json:"total"`
	OpenCount      int        `json:"openCount"`
	WorkerLimit    int        `json:"workerLimit,omitempty"`
	PortsPerBatch  int        `json:"portsPerBatch,omitempty"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	FinishedAt     *time.Time `json:"finishedAt,omitempty"`
	Error          string     `json:"error,omitempty"`
	Mode           string     `json:"mode,omitempty"`
	ClosedCount    int        `json:"closedCount"`
	UncertainCount int        `json:"uncertainCount"`
	TimeoutCount   int        `json:"timeoutCount"`
	ResponseCount  int        `json:"responseCount"`
}
type EmailRequest struct {
	Recipients string `json:"recipients"`
}
type SMTPTestRequest struct {
	Recipient string `json:"recipient"`
}
type ScanPayload struct {
	Status  ScanStatus `json:"status"`
	Results []OpenPort `json:"results"`
}
type Dashboard struct {
	Statuses []ScanStatus `json:"statuses"`
	Results  []OpenPort   `json:"results"`
	Warning  string       `json:"warning,omitempty"`
}
type RemoteClient struct {
	BaseURL, Token string
	Client         *http.Client
}

func (r *RemoteClient) request(method, endpoint string, input any, output any) error {
	var body io.Reader
	if input != nil {
		b, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, strings.TrimRight(r.BaseURL, "/")+endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-Scanner-Token", r.Token)
	req.Header.Set("Content-Type", "application/json")
	response, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return fmt.Errorf("uzak tarayici HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(b)))
	}
	if output != nil {
		return json.NewDecoder(response.Body).Decode(output)
	}
	return nil
}

type Manager struct {
	mu         sync.RWMutex
	feed       Feed
	feedFile   string
	targets    []Target
	results    []OpenPort
	status     ScanStatus
	name       string
	resultFile string
	cancel     context.CancelFunc
	subs       map[chan ScanEvent]struct{}
}

func newManager(feed Feed, scannerName, feedFile, resultFile string) *Manager {
	if scannerName == "" {
		scannerName = "Turkiye"
	}
	m := &Manager{feed: feed, feedFile: feedFile, name: scannerName, resultFile: resultFile, results: make([]OpenPort, 0), subs: make(map[chan ScanEvent]struct{})}
	m.status.ScannerName = scannerName
	m.targets = flatten(feed)
	m.restoreResults()
	return m
}
func (m *Manager) restoreResults() {
	b, err := os.ReadFile(m.resultFile)
	if err != nil {
		return
	}
	var results []OpenPort
	if json.Unmarshal(b, &results) == nil {
		m.results = results
		m.status.OpenCount = len(results)
	}
}
func (m *Manager) persistResults() {
	_, results := m.snapshot()
	if err := os.MkdirAll(filepath.Dir(m.resultFile), 0755); err != nil {
		return
	}
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return
	}
	temporary := m.resultFile + ".tmp"
	if os.WriteFile(temporary, b, 0600) == nil {
		_ = os.Rename(temporary, m.resultFile)
	}
}
func (m *Manager) snapshotFeed() (Feed, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.feed, len(m.targets)
}
func (m *Manager) updateFeed(feed Feed) error {
	for _, device := range feed.Devices {
		for _, ip := range device.IPAddresses {
			if net.ParseIP(strings.TrimSpace(ip)) == nil {
				return fmt.Errorf("gecersiz IP adresi: %s", ip)
			}
		}
	}
	targets := flatten(feed)
	if len(targets) == 0 {
		return errors.New("en az bir gecerli IP ekleyin")
	}
	m.mu.Lock()
	if m.status.Running {
		m.mu.Unlock()
		return errors.New("tarama devam ederken IP listesi degistirilemez")
	}
	m.feed, m.targets = feed, targets
	m.mu.Unlock()
	b, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return err
	}
	temporary := m.feedFile + ".tmp"
	if err := os.WriteFile(temporary, b, 0600); err != nil {
		return err
	}
	return os.Rename(temporary, m.feedFile)
}
func flatten(f Feed) []Target {
	var ts []Target
	for _, d := range f.Devices {
		for _, ip := range d.IPAddresses {
			if net.ParseIP(ip) != nil {
				ts = append(ts, Target{ip, d.DeviceName, d.CustomerName, d.DealerName})
			}
		}
	}
	return ts
}
func (m *Manager) emit(e ScanEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for ch := range m.subs {
		select {
		case ch <- e:
		default:
		}
	}
}
func (m *Manager) subscribe() (chan ScanEvent, func()) {
	ch := make(chan ScanEvent, 64)
	m.mu.Lock()
	m.subs[ch] = struct{}{}
	m.mu.Unlock()
	return ch, func() { m.mu.Lock(); delete(m.subs, ch); close(ch); m.mu.Unlock() }
}
func (m *Manager) snapshot() (ScanStatus, []OpenPort) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.status
	r := append([]OpenPort{}, m.results...)
	return s, r
}
func (m *Manager) payload() ScanPayload {
	status, results := m.snapshot()
	return ScanPayload{Status: status, Results: results}
}

func dashboard(local *Manager, remote *RemoteClient) Dashboard {
	localPayload := local.payload()
	output := Dashboard{Statuses: []ScanStatus{localPayload.Status}, Results: localPayload.Results}
	if remote == nil {
		return output
	}
	var remotePayload ScanPayload
	if err := remote.request(http.MethodGet, "/api/internal/scan", nil, &remotePayload); err != nil {
		output.Warning = "Yurtdisi tarayiciya ulasilamadi: " + err.Error()
		return output
	}
	output.Statuses = append(output.Statuses, remotePayload.Status)
	output.Results = append(output.Results, remotePayload.Results...)
	sort.Slice(output.Results, func(i, j int) bool {
		if output.Results[i].Source == output.Results[j].Source {
			if output.Results[i].IP == output.Results[j].IP {
				return output.Results[i].Port < output.Results[j].Port
			}
			return output.Results[i].IP < output.Results[j].IP
		}
		return output.Results[i].Source < output.Results[j].Source
	})
	return output
}
func (m *Manager) start(req ScanRequest) error {
	if len(req.Ports) == 0 {
		if req.StartPort == 0 {
			req.StartPort = 1
		}
		if req.EndPort == 0 {
			req.EndPort = 65535
		}
		for port := req.StartPort; port <= req.EndPort; port++ {
			req.Ports = append(req.Ports, port)
		}
	}
	if req.TimeoutMS == 0 {
		req.TimeoutMS = 250
	}
	if req.Workers == 0 {
		req.Workers = 4000
	}
	req.Mode = strings.ToLower(strings.TrimSpace(req.Mode))
	if req.Mode == "" {
		req.Mode = "normal"
	}
	if req.Mode != "normal" && req.Mode != "fast" {
		return errors.New("ge?ersiz tarama modu")
	}
	if req.Mode == "fast" {
		req.TimeoutMS = 140
		req.Workers = 5000
		if req.PortsPerBatch == 0 {
			req.PortsPerBatch = 400
		}
	}
	unique := make(map[int]struct{}, len(req.Ports))
	ports := make([]int, 0, len(req.Ports))
	for _, port := range req.Ports {
		if port < 1 || port > 65535 {
			return errors.New("port 1 ile 65535 arasynda olmali")
		}
		if _, exists := unique[port]; !exists {
			unique[port] = struct{}{}
			ports = append(ports, port)
		}
	}
	sort.Ints(ports)
	req.Ports = ports
	if len(req.Ports) == 0 || req.TimeoutMS < 25 || req.TimeoutMS > 30000 || req.Workers < 1 || req.Workers > 5000 || req.PortsPerBatch < 0 || req.PortsPerBatch > 1000 {
		return errors.New("gecersiz tarama ayarlari")
	}
	m.mu.Lock()
	if m.status.Running {
		m.mu.Unlock()
		return errors.New("tarama zaten calisiyor")
	}
	if len(m.targets) == 0 {
		m.mu.Unlock()
		return errors.New("gecerli IP bulunamadi")
	}
	portsPerBatch := req.PortsPerBatch
	if portsPerBatch == 0 {
		// Batch, worker limitini asmamalidir. Aksi halde sondaki birkac is
		// ikinci dalgaya kalir ve batch suresi gereksiz yere ikiye katlanir.
		portsPerBatch = req.Workers / len(m.targets)
		if portsPerBatch < 1 {
			portsPerBatch = 1
		}
	}
	req.PortsPerBatch = portsPerBatch
	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now()
	m.cancel = cancel
	m.results = nil
	m.status = ScanStatus{ScannerName: m.name, Running: true, CurrentPort: req.Ports[0], Total: len(req.Ports) * len(m.targets), WorkerLimit: req.Workers, PortsPerBatch: req.PortsPerBatch, Mode: req.Mode, StartedAt: &now}
	targets := append([]Target(nil), m.targets...)
	m.mu.Unlock()
	go m.run(ctx, req, targets)
	return nil
}
func (m *Manager) run(ctx context.Context, req ScanRequest, targets []Target) {
	defer m.finishScan()
	workerCount := minInt(req.Workers, len(req.Ports)*len(targets))
	if workerCount < 1 {
		return
	}
	perIPLimit := minInt(96, maxInt(4, req.Workers/maxInt(1, len(targets))))
	if req.Mode == "fast" {
		perIPLimit = minInt(192, perIPLimit)
	}
	limits := make(map[string]chan struct{}, len(targets))
	for _, target := range targets {
		if _, ok := limits[target.IP]; !ok {
			limits[target.IP] = make(chan struct{}, perIPLimit)
		}
	}
	queue := make(chan BatchJob, workerCount)
	fastTimeout := time.Duration(req.TimeoutMS) * time.Millisecond
	if req.Mode == "fast" {
		fastTimeout = 140 * time.Millisecond
	}
	confirmTimeout := maxDuration(time.Second, fastTimeout*4)
	if req.Mode == "fast" {
		confirmTimeout = 450 * time.Millisecond
	}
	record := func(job ScanJob, outcome string) {
		m.mu.Lock()
		m.status.CurrentPort, m.status.Completed = job.Port, m.status.Completed+1
		m.status.ResponseCount++
		completed, total := m.status.Completed, m.status.Total
		if outcome == "open" {
			r := OpenPort{Source: m.name, IP: job.Target.IP, DeviceName: job.Target.DeviceName, CustomerName: job.Target.CustomerName, DealerName: job.Target.DealerName, Port: job.Port, FoundAt: time.Now()}
			m.results = append(m.results, r)
			m.status.OpenCount++
			m.mu.Unlock()
			m.emit(ScanEvent{Type: "result", Port: job.Port, Completed: completed, Total: total, Result: &r})
		} else {
			if outcome == "uncertain" {
				m.status.UncertainCount++
				m.status.TimeoutCount++
			} else {
				m.status.ClosedCount++
			}
			m.mu.Unlock()
		}
		m.emit(ScanEvent{Type: "progress", Port: job.Port, Completed: completed, Total: total})
	}
	resolveUncertain := func(job ScanJob, outcome string) {
		m.mu.Lock()
		m.status.CurrentPort = job.Port
		if m.status.UncertainCount > 0 {
			m.status.UncertainCount--
		}
		if outcome == "open" {
			r := OpenPort{Source: m.name, IP: job.Target.IP, DeviceName: job.Target.DeviceName, CustomerName: job.Target.CustomerName, DealerName: job.Target.DealerName, Port: job.Port, FoundAt: time.Now()}
			m.results = append(m.results, r)
			m.status.OpenCount++
			completed, total := m.status.Completed, m.status.Total
			m.mu.Unlock()
			m.emit(ScanEvent{Type: "result", Port: job.Port, Completed: completed, Total: total, Result: &r})
			return
		}
		m.status.ClosedCount++
		m.mu.Unlock()
	}
	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range queue {
				limit := limits[job.Target.IP]
				select {
				case limit <- struct{}{}:
				case <-ctx.Done():
					job.done.Done()
					continue
				}
				if job.confirm {
					if req.Mode == "fast" {
						resolveUncertain(job.ScanJob, dialOutcome(ctx, job.Target.IP, job.Port, confirmTimeout, true))
					} else {
						record(job.ScanJob, dialOutcome(ctx, job.Target.IP, job.Port, confirmTimeout, true))
					}
				} else {
					outcome := dialOutcome(ctx, job.Target.IP, job.Port, fastTimeout, false)
					if outcome == "uncertain" {
						if req.Mode == "fast" {
							record(job.ScanJob, "uncertain")
						}
						job.retryMu.Lock()
						*job.retryJobs = append(*job.retryJobs, job.ScanJob)
						job.retryMu.Unlock()
					} else {
						record(job.ScanJob, outcome)
					}
				}
				<-limit
				job.done.Done()
			}
		}()
	}
	defer func() { close(queue); workers.Wait() }()
	dispatch := func(jobs []ScanJob, confirm bool, retries *[]ScanJob, retriesMu *sync.Mutex) bool {
		var done sync.WaitGroup
		for _, scanJob := range jobs {
			done.Add(1)
			select {
			case queue <- BatchJob{ScanJob: scanJob, done: &done, confirm: confirm, retryJobs: retries, retryMu: retriesMu}:
			case <-ctx.Done():
				done.Done()
				done.Wait()
				return false
			}
		}
		done.Wait()
		return ctx.Err() == nil
	}
	var deferredRetries []ScanJob
	for first := 0; first < len(req.Ports); first += req.PortsPerBatch {
		last := minInt(len(req.Ports), first+req.PortsPerBatch)
		jobs := make([]ScanJob, 0, (last-first)*len(targets))
		for _, port := range req.Ports[first:last] {
			for _, target := range targets {
				jobs = append(jobs, ScanJob{Target: target, Port: port})
			}
		}
		var retries []ScanJob
		var retriesMu sync.Mutex
		if !dispatch(jobs, false, &retries, &retriesMu) {
			m.emit(ScanEvent{Type: "cancelled", Message: "Tarama durduruldu"})
			return
		}
		if req.Mode == "fast" {
			deferredRetries = append(deferredRetries, retries...)
		} else if len(retries) > 0 && !dispatch(retries, true, nil, nil) {
			m.emit(ScanEvent{Type: "cancelled", Message: "Tarama durduruldu"})
			return
		}
	}
	if req.Mode == "fast" && len(deferredRetries) > 0 {
		_ = dispatch(deferredRetries, true, nil, nil)
	}
}
func (m *Manager) finishScan() {
	m.mu.Lock()
	m.status.Running = false
	m.cancel = nil
	now := time.Now()
	m.status.FinishedAt = &now
	m.mu.Unlock()
	m.persistResults()
	m.emit(ScanEvent{Type: "finished"})
}
func dialOutcome(ctx context.Context, ip string, port int, timeout time.Duration, confirmed bool) string {
	c, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err == nil {
		_ = c.Close()
		return "open"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() && ctx.Err() == nil {
		if confirmed {
			return "uncertain"
		}
		return "uncertain"
	}
	return "closed"
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func (m *Manager) stop() {
	m.mu.RLock()
	cancel := m.cancel
	m.mu.RUnlock()
	if cancel != nil {
		cancel()
	}
}

func parseRecipients(recipients string) ([]string, error) {
	var to []string
	for _, address := range strings.Split(recipients, ",") {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}
		if _, err := mail.ParseAddress(address); err != nil {
			return nil, fmt.Errorf("ge\u00e7ersiz al\u0131c\u0131 e-posta adresi: %s", address)
		}
		to = append(to, address)
	}
	if len(to) == 0 {
		return nil, errors.New("en az bir al\u0131c\u0131 e-posta adresi girin")
	}
	return to, nil
}

func smtpMessage(from string, to []string, subject, body string) []byte {
	return []byte("From: " + from + "\r\nTo: " + strings.Join(to, ", ") + "\r\nSubject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n" + body)
}

func sendSMTPReport(to []string, subject, body string) error {
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if from == "" {
		from = "gokhangokcenn@gmail.com"
	}
	message := smtpMessage(from, to, subject, body)
	internalHost := strings.TrimSpace(os.Getenv("INTERNAL_SMTP_HOST"))
	if internalHost == "" {
		internalHost = "172.16.0.6"
	}
	internalPort := strings.TrimSpace(os.Getenv("INTERNAL_SMTP_PORT"))
	if internalPort == "" {
		internalPort = "25"
	}
	if err := smtp.SendMail(net.JoinHostPort(internalHost, internalPort), nil, from, to, message); err == nil {
		return nil
	} else {
		gmailUser := strings.TrimSpace(os.Getenv("GMAIL_SMTP_USERNAME"))
		if gmailUser == "" {
			gmailUser = from
		}
		gmailPassword := strings.ReplaceAll(strings.TrimSpace(os.Getenv("GMAIL_SMTP_APP_PASSWORD")), " ", "")
		if gmailPassword == "" {
			return fmt.Errorf("dahili SMTP ba\u015far\u0131s\u0131z: %v; Gmail yede\u011fi i\u00e7in GMAIL_SMTP_APP_PASSWORD tan\u0131ml\u0131 de\u011fil", err)
		}
		if gmailErr := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", gmailUser, gmailPassword, "smtp.gmail.com"), from, to, message); gmailErr != nil {
			return fmt.Errorf("dahili SMTP ba\u015far\u0131s\u0131z: %v; Gmail yede\u011fi ba\u015far\u0131s\u0131z: %w", err, gmailErr)
		}
		return nil
	}
}

func (m *Manager) emailReport(recipients string, results []OpenPort) error {
	to, err := parseRecipients(recipients)
	if err != nil {
		return err
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].IP == results[j].IP {
			return results[i].Port < results[j].Port
		}
		return results[i].IP < results[j].IP
	})
	var body strings.Builder
	body.WriteString("Port Multi Scanner\nA\u00e7\u0131k Port Raporu\n\n")
	body.WriteString(fmt.Sprintf("Toplam a\u00e7\u0131k port: %d\n\n", len(results)))
	for i := 0; i < len(results); {
		r := results[i]
		var ports []string
		j := i
		for j < len(results) && results[j].IP == r.IP {
			ports = append(ports, strconv.Itoa(results[j].Port))
			j++
		}
		body.WriteString(fmt.Sprintf("IP: %s\nSunucu: %s\nM\u00fc\u015fteri: %s\nBayi: %s\nA\u00e7\u0131k portlar (%d): %s\n\n", r.IP, emptyLabel(r.DeviceName), emptyLabel(r.CustomerName), emptyLabel(r.DealerName), len(ports), strings.Join(ports, ", ")))
		i = j
	}
	if len(results) == 0 {
		body.WriteString("Bu taramada a\u00e7\u0131k TCP port bulunamad\u0131.\n")
	}
	return sendSMTPReport(to, "Port Multi Scanner - A\u00e7\u0131k Port Raporu", body.String())
}

func emptyLabel(value string) string {
	if strings.TrimSpace(value) == "" {
		return "\u2014"
	}
	return value
}
func (m *Manager) smtpTest(recipient string) error {
	to, err := parseRecipients(recipient)
	if err != nil {
		return err
	}
	return sendSMTPReport(to, "Port Multi Scanner - SMTP Testi", "SMTP ayarlar\u0131 do\u011fruland\u0131.\n")
}
func loadFeed(path string) (Feed, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return Feed{}, e
	}
	var f Feed
	e = json.Unmarshal(b, &f)
	return f, e
}

func projectFile(name string) string {
	for _, path := range []string{name, filepath.Join("..", name)} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return name
}

func loadDotEnv(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pair := strings.SplitN(line, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key, value := strings.TrimSpace(pair[0]), strings.TrimSpace(pair[1])
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}
func main() {
	loadDotEnv(projectFile(".env"))
	feed, err := loadFeed(projectFile("ip_feed.json"))
	if err != nil {
		panic(fmt.Sprintf("ip_feed.json okunamadi: %v", err))
	}
	resultsFile := os.Getenv("RESULTS_FILE")
	if resultsFile == "" {
		resultsFile = projectFile("data/results.json")
	}
	manager := newManager(feed, os.Getenv("SCANNER_NAME"), projectFile("ip_feed.json"), resultsFile)
	var remote *RemoteClient
	if remoteURL, remoteToken := os.Getenv("REMOTE_SCANNER_URL"), os.Getenv("REMOTE_SCANNER_TOKEN"); remoteURL != "" && remoteToken != "" {
		remote = &RemoteClient{BaseURL: remoteURL, Token: remoteToken, Client: &http.Client{Timeout: 8 * time.Second}}
	}
	internalToken := os.Getenv("SCANNER_API_TOKEN")
	internalAuthorized := func(c *fiber.Ctx) bool { return internalToken != "" && c.Get("X-Scanner-Token") == internalToken }
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, e error) error {
		code := fiber.StatusInternalServerError
		var fiberError *fiber.Error
		if errors.As(e, &fiberError) {
			code = fiberError.Code
		}
		return c.Status(code).JSON(fiber.Map{"error": e.Error()})
	}})
	api := app.Group("/api")
	api.Get("/feed", func(c *fiber.Ctx) error {
		feed, targetCount := manager.snapshotFeed()
		return c.JSON(fiber.Map{"devices": feed.Devices, "targetCount": targetCount})
	})
	api.Put("/feed", func(c *fiber.Ctx) error {
		var feed Feed
		if err := c.BodyParser(&feed); err != nil {
			return fiber.NewError(400, "JSON ge\u00e7ersiz")
		}
		if err := manager.updateFeed(feed); err != nil {
			return fiber.NewError(409, err.Error())
		}
		_, targetCount := manager.snapshotFeed()
		return c.JSON(fiber.Map{"message": "IP listesi kaydedildi", "targetCount": targetCount})
	})
	api.Get("/scan", func(c *fiber.Ctx) error {
		s, r := manager.snapshot()
		return c.JSON(fiber.Map{"status": s, "results": r})
	})
	api.Get("/dashboard", func(c *fiber.Ctx) error { return c.JSON(dashboard(manager, remote)) })
	api.Get("/internal/scan", func(c *fiber.Ctx) error {
		if !internalAuthorized(c) {
			return fiber.NewError(401, "yetkisiz uzak tarayici istegi")
		}
		return c.JSON(manager.payload())
	})
	api.Post("/scan", func(c *fiber.Ctx) error {
		var r ScanRequest
		if e := c.BodyParser(&r); e != nil {
			return fiber.NewError(400, "JSON ge\u00e7ersiz")
		}
		if e := manager.start(r); e != nil {
			return fiber.NewError(409, e.Error())
		}
		if remote != nil {
			go func() { _ = remote.request(http.MethodPost, "/api/internal/scan", r, nil) }()
		}
		return c.Status(202).JSON(fiber.Map{"message": "Tarama baslatildi"})
	})
	api.Post("/internal/scan", func(c *fiber.Ctx) error {
		if !internalAuthorized(c) {
			return fiber.NewError(401, "yetkisiz uzak tarayici istegi")
		}
		var request ScanRequest
		if err := c.BodyParser(&request); err != nil {
			return fiber.NewError(400, "JSON ge\u00e7ersiz")
		}
		if err := manager.start(request); err != nil {
			return fiber.NewError(409, err.Error())
		}
		return c.SendStatus(202)
	})
	api.Delete("/internal/scan", func(c *fiber.Ctx) error {
		if !internalAuthorized(c) {
			return fiber.NewError(401, "yetkisiz uzak tarayici istegi")
		}
		manager.stop()
		return c.SendStatus(204)
	})
	api.Delete("/scan", func(c *fiber.Ctx) error {
		manager.stop()
		if remote != nil {
			go func() { _ = remote.request(http.MethodDelete, "/api/internal/scan", nil, nil) }()
		}
		return c.SendStatus(204)
	})
	api.Post("/smtp-test", func(c *fiber.Ctx) error {
		var request SMTPTestRequest
		if err := c.BodyParser(&request); err != nil {
			return fiber.NewError(400, "JSON ge\u00e7ersiz")
		}
		if err := manager.smtpTest(request.Recipient); err != nil {
			return fiber.NewError(400, err.Error())
		}
		return c.JSON(fiber.Map{"message": "SMTP test e-postas\u0131 g\u00f6nderildi"})
	})
	api.Post("/email-report", func(c *fiber.Ctx) error {
		var request EmailRequest
		if err := c.BodyParser(&request); err != nil {
			return fiber.NewError(400, "JSON ge\u00e7ersiz")
		}
		if err := manager.emailReport(request.Recipients, dashboard(manager, remote).Results); err != nil {
			return fiber.NewError(400, err.Error())
		}
		return c.JSON(fiber.Map{"message": "Rapor e-posta ile gonderildi"})
	})
	api.Get("/events", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		ch, unsub := manager.subscribe()
		defer unsub()
		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			for e := range ch {
				b, _ := json.Marshal(e)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Type, b)
				w.Flush()
			}
		})
		return nil
	})
	app.Use("/", filesystem.New(filesystem.Config{Root: http.Dir(projectFile("frontend/dist")), NotFoundFile: "index.html"}))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	panic(app.Listen(":" + port))
}
