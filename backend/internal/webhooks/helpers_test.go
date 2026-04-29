package webhooks

import (
	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
)

func stubHook(url, secret string) db.Webhook {
	return db.Webhook{
		ID:     uuid.New(),
		Url:    url,
		Secret: secret,
		Active: true,
	}
}
