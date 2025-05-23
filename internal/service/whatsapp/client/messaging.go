package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/gwenziro/bot-notify/internal/utils"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// SendMessage mengirim pesan teks ke nomor atau grup tertentu
func (c *Client) SendMessage(recipient types.JID, message string) error {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return errors.New("klien WhatsApp belum terhubung")
	}

	c.logger.WithFields(utils.Fields{
		"to":             recipient.String(),
		"message_length": len(message),
	}).Info("Mengirim pesan")

	// Update aktivitas
	c.UpdateLastActivity()

	// Kirim pesan
	_, err := c.waClient.SendMessage(context.Background(), recipient, &waProto.Message{
		Conversation: &message,
	})

	if err != nil {
		return fmt.Errorf("gagal mengirim pesan: %w", err)
	}

	return nil
}

// SendFormattedMessage mengirim pesan dengan format khusus (bold, italic, dll)
func (c *Client) SendFormattedMessage(recipient types.JID, message string) error {
	if c.waClient == nil || !c.connectionState.IsConnected {
		return errors.New("klien WhatsApp belum terhubung")
	}

	c.logger.WithFields(utils.Fields{
		"to":             recipient.String(),
		"message_length": len(message),
		"type":           "formatted",
	}).Info("Mengirim pesan terformat")

	// Update aktivitas
	c.UpdateLastActivity()

	// Konversi ke ExtendedTextMessage untuk dukungan format
	_, err := c.waClient.SendMessage(context.Background(), recipient, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: &message,
			// Bisa ditambahkan opsi pemformatan lainnya
		},
	})

	if err != nil {
		return fmt.Errorf("gagal mengirim pesan terformat: %w", err)
	}

	return nil
}

// BroadcastMessage mengirim pesan ke beberapa penerima sekaligus
func (c *Client) BroadcastMessage(recipients []types.JID, message string) map[string]error {
	results := make(map[string]error)

	for _, recipient := range recipients {
		err := c.SendMessage(recipient, message)
		results[recipient.String()] = err
	}

	return results
}
