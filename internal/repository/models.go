package repository

// errors

// usersTable
const (
	usersTable = "users"

	idColumn       = "id"
	usernameColumn = "username"
	passwordColumn = "password"
)

// messagesTable
const (
	messagesTable = "messages"

	senderIdColumn   = "sender_id"
	receiverIdColumn = "receiver_id"
	textColumn       = "text"
	timeToSendColumn = "time_to_send"
)
