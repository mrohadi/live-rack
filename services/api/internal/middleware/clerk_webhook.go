package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	svix "github.com/svix/svix-webhooks/go"
)

// ClerkWebhookEvent covers the subset of events we handle.
type ClerkWebhookEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ClerkOrgCreatedData holds org fields from a Clerk organization.created event.
type ClerkOrgCreatedData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ClerkUserCreatedData holds user fields from a Clerk user.created event.
type ClerkUserCreatedData struct {
	ID             string `json:"id"`
	EmailAddresses []struct {
		EmailAddress string `json:"email_address"`
	} `json:"email_addresses"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ImageURL       string `json:"image_url"`
	OrganizationID string `json:"organization_id"`
}

// WebhookProvisioner provisions orgs and users from Clerk events.
type WebhookProvisioner interface {
	ProvisionOrg(orgID, name string) error
	ProvisionUser(clerkUserID, clerkOrgID, email, displayName, avatarURL string) error
}

// ClerkWebhookHandler verifies and dispatches Clerk webhook events.
type ClerkWebhookHandler struct {
	wh          *svix.Webhook
	provisioner WebhookProvisioner
}

// NewClerkWebhookHandler constructs a handler that verifies Clerk webhook signatures.
func NewClerkWebhookHandler(signingSecret string, p WebhookProvisioner) (*ClerkWebhookHandler, error) {
	wh, err := svix.NewWebhook(signingSecret)
	if err != nil {
		return nil, fmt.Errorf("clerk webhook: init svix: %w", err)
	}
	return &ClerkWebhookHandler{wh: wh, provisioner: p}, nil
}

func (h *ClerkWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.wh.Verify(body, r.Header); err != nil {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var evt ClerkWebhookEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}

	switch evt.Type {
	case "organization.created":
		var d ClerkOrgCreatedData
		if err := json.Unmarshal(evt.Data, &d); err == nil {
			_ = h.provisioner.ProvisionOrg(d.ID, d.Name)
		}
	case "user.created":
		var d ClerkUserCreatedData
		if err := json.Unmarshal(evt.Data, &d); err == nil {
			email := ""
			if len(d.EmailAddresses) > 0 {
				email = d.EmailAddresses[0].EmailAddress
			}
			displayName := d.FirstName + " " + d.LastName
			_ = h.provisioner.ProvisionUser(d.ID, d.OrganizationID, email, displayName, d.ImageURL)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
