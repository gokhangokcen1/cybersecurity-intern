package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gokhangokcen1/subnet-backend/crafter"
	"github.com/gokhangokcen1/subnet-backend/models"
)

func PortCraftHandler(c fiber.Ctx) error {
	req := new(models.PacketCraftRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.PacketCraftResponse{
			Success: false,
			Message: "İstek parametreleri okunamadı.",
		})
	}

	if req.Protocol == "" || req.SrcIP == "" || req.DstIP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.PacketCraftResponse{
			Success: false,
			Message: "Protokol, Kaynak IP ve Hedef IP alanları zorunludur.",
		})
	}

	err := crafter.SendCraftedPacket(*req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.PacketCraftResponse{
			Success: false,
			Message: fmt.Sprintf("paket durumu: %s", err.Error()),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.PacketCraftResponse{
		Success: true,
		Message: "Paket başarıyla oluşturuldu!",
	})
}
