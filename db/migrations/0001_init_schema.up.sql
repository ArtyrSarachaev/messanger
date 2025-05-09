CREATE TABLE IF NOT EXISTS users
(
    -- Id, как правило, делается uuid, если тебе не нужен int,
    -- а у тебя, как вижу, необходимости использовать int нет.
    -- Для сообщений - другое дело, там ты по этому ключу ещё и сортировать по идее должен
    id       SERIAL PRIMARY KEY,
    -- Varchar не занимает меньше места в памяти, он просто проверяет длину и выкинет ошибку,
    -- если она превышает заданную. Так что если хочешь пользаку показать, что слишком длинное имя
    -- лучше это сделать в бизнес-логике
    -- Поэтому можно просто использовать text
    username VARCHAR(50)  NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
    -- Также хорошая практика добавлять created_at и updated_at колонки.
    -- Дополнительно можно запилить deleted_at и проставлять его, если запись удаляется.
    -- Но это история для таблиц с дохуилионом записей,
    -- т.к. удаление строки гораздо более затратная операция, чем её обновление

    -- ещё было бы славно добавить индекс по username, если шаришь за такое
    -- потому что у тебя часто вызывается GetByName и это здорово оптимизирует запрос
);

CREATE TABLE IF NOT EXISTS messages
(
    id           SERIAL PRIMARY KEY,
    sender_id    INT  NOT NULL,
    receiver_id  INT  NOT NULL,
    text         TEXT NOT NULL,
    time_to_send INT8 DEFAULT EXTRACT(EPOCH FROM NOW()), -- а почему не timestamp а int8?
    FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (receiver_id) REFERENCES users (id) ON DELETE CASCADE
);