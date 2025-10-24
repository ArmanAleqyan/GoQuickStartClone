package email

import (
	"fmt"
	"log"
	"sync"
)

type EmailService struct {
	fromEmail string
	queue     chan EmailJob
	wg        sync.WaitGroup
}

type EmailJob struct {
	To      string
	Subject string
	Body    string
}

func NewEmailService(fromEmail string) *EmailService {
	service := &EmailService{
		fromEmail: fromEmail,
		queue:     make(chan EmailJob, 1000), // Buffered channel for 1000 emails
	}

	// Start background workers
	service.startWorkers(5) // 5 concurrent email workers

	return service
}

// startWorkers spawns goroutines to process email queue
func (s *EmailService) startWorkers(count int) {
	for i := 0; i < count; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

// worker processes emails from the queue
func (s *EmailService) worker(id int) {
	defer s.wg.Done()
	log.Printf("[EmailService] Worker %d started", id)

	for job := range s.queue {
		s.sendEmail(job)
	}

	log.Printf("[EmailService] Worker %d stopped", id)
}

// sendEmail performs the actual email sending
func (s *EmailService) sendEmail(job EmailJob) {
	// In production, use actual SMTP or email service
	log.Printf("=== EMAIL SENT ===")
	log.Printf("From: %s", s.fromEmail)
	log.Printf("To: %s", job.To)
	log.Printf("Subject: %s", job.Subject)
	log.Printf("Body: %s", job.Body)
	log.Printf("==================")

	fmt.Printf("\n📧 Email sent to: %s\n", job.To)
	fmt.Printf("📝 Subject: %s\n\n", job.Subject)
}

// queueEmail adds an email to the queue for async sending
func (s *EmailService) queueEmail(to, subject, body string) {
	select {
	case s.queue <- EmailJob{To: to, Subject: subject, Body: body}:
		log.Printf("[EmailService] Email queued for %s", to)
	default:
		log.Printf("[EmailService] Email queue full, dropping email to %s", to)
	}
}

// Shutdown gracefully shuts down the email service
func (s *EmailService) Shutdown() {
	log.Printf("[EmailService] Shutting down...")
	close(s.queue)
	s.wg.Wait()
	log.Printf("[EmailService] Shutdown complete")
}

// SendPasswordResetEmail отправляет email со ссылкой для сброса пароля асинхронно
// В production версии здесь должна быть интеграция с SMTP сервером или сервисом типа SendGrid
func (s *EmailService) SendPasswordResetEmail(toEmail, resetToken, resetURL string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
Password Reset Request

Hello,

You requested to reset your password. Please click the link below to reset your password:

%s?token=%s

This link will expire in 1 hour.

If you didn't request this, please ignore this email.

Token: %s

---
This email was sent automatically. Please do not reply.
`, resetURL, resetToken, resetToken)

	// Queue email for async sending
	s.queueEmail(toEmail, subject, body)

	// В dev режиме также логируем для удобства
	fmt.Printf("\n📧 Email queued for: %s\n", toEmail)
	fmt.Printf("🔗 Reset link: %s?token=%s\n\n", resetURL, resetToken)

	return nil
}

// SendWelcomeEmail отправляет приветственное письмо новому пользователю асинхронно
func (s *EmailService) SendWelcomeEmail(toEmail, firstName string) error {
	subject := "Welcome to IronNode!"
	body := fmt.Sprintf(`
Welcome to IronNode!

Hello %s,

Thank you for registering with IronNode!

You now have access to our blockchain node infrastructure platform. You can start making requests to various blockchain networks immediately.

Getting Started:
1. Log in to your account
2. Create an API key in your dashboard
3. Start making blockchain requests

If you have any questions, feel free to reach out to our support team.

Best regards,
IronNode Team

---
This email was sent automatically. Please do not reply.
`, firstName)

	// Queue email for async sending
	s.queueEmail(toEmail, subject, body)

	fmt.Printf("📧 Welcome email queued for: %s\n", toEmail)

	return nil
}

// SendPasswordChangedEmail уведомляет о смене пароля асинхронно
func (s *EmailService) SendPasswordChangedEmail(toEmail string) error {
	subject := "Password Changed Successfully"
	body := `
Password Changed Successfully

Hello,

Your password has been changed successfully.

If you didn't make this change, please contact our support team immediately.

Time: ` + fmt.Sprintf("%v", "just now") + `

Best regards,
IronNode Team

---
This email was sent automatically. Please do not reply.
`

	// Queue email for async sending
	s.queueEmail(toEmail, subject, body)

	fmt.Printf("📧 Password changed notification queued for: %s\n", toEmail)

	return nil
}
