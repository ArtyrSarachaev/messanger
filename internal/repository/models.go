package repository

// errors
const (
	// В логи выкидывать запрос не совсем корректно
	// Вдруг там будут какие-то приватные данные пользователя или ещё что-то важное
	cantExecQueryWithError = "cant exec query '%s', with error %v"
)

// usersTable
const (
	usersTable = "users" // Про имя таблицы уже писал в entity/user.go

	/**
	Вот эти названия колонов в константы выносить это тоже дрочь лютая и ненужная
	настолько упарываться - только время терять
	*/
	idColumn       = "id"
	usernameColumn = "username"
	passwordColumn = "password"
)

// messagesTable
const (
	messagesTable = "messages"
)
