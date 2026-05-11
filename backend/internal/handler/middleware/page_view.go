package middleware

import (
	"fmt"
	"log"
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/click"
	auth "dd-prediction-api/pkg/jwt"

	"github.com/gofiber/fiber/v3"
)

func PageViewMiddleware(c fiber.Ctx) error {
	err := c.Next()
	if err != nil {
		return err
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return nil
	}

	if !model.ShouldTrackActivity(c.Context()) {
		log.Printf("Not tracking activity for path: %s", c.Path())
		return nil
	}

	click.GetActivityBuffer().Add(model.UserActivity{
		UserID:        claims.UserID,
		ActivityType:  model.ActivityPageView,
		CreatedAt:     time.Now().UTC(),
		ActivityCount: 1,
		Metadata:      fmt.Sprintf(`{"path":"%s","method":"%s"}`, c.Path(), c.Method()),
	})

	return nil
}
