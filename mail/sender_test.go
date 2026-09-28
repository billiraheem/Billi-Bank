package mail

import (
	"testing"

	"github.com/billiraheem/Billi-Bank/utils"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	config, err := utils.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddr, config.EmailSenderPassword)

	subject := "A test email"
	content := `
		<h1>Hello World</h1>
		<p>This is a test message from Balikis</p>
	`
	to := []string{config.EmailReceiverAddr}
	attachedFiles := []string{"../start.sh"}

	err = sender.Send(subject, content, to, nil, nil, attachedFiles)
	require.NoError(t, err)
}