package services

import (
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/webhook"
)

type StripeService struct {
	secretKey string
}

func NewStripeService(secretKey string) *StripeService {
	stripe.Key = secretKey // Required by stripe-go SDK initialization
	return &StripeService{
		secretKey: secretKey,
	}
}

// Create Payment Session
func (s *StripeService) CreatePaymentSession(userID string, amount int64, returnURL string) (string, error) {
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Image Pack"),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(returnURL + "/payment-success"),
		CancelURL:  stripe.String(returnURL + "/payment-canceled"),
		ClientReferenceID: stripe.String(userID),
	}

	sess, err := session.New(params)
	if err != nil {
		return "", err
	}

	return sess.URL, nil
}

// ValidateWebhookSignature validates the webhook signature
func (s *StripeService) ValidateWebhookSignature(body []byte, sig string, endpointSecret string) ([]byte, error) {
	event, err := webhook.ConstructEvent(body, sig, endpointSecret)
	if err != nil {
		return nil, err
	}

	return event.Data.Raw, nil
}
