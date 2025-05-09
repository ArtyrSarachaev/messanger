* Нет файлика .gitignore
* Нет файлика Makefile, поэтому непонятно как проект запустить, 
какие скрипты нужно вызвать
* db_requests → db/migrations. И вообще нужно почитать про миграции. Вот с этого можешь начать https://habr.com/ru/articles/780280/
  * Миграции это запросы для БД, которые приводят её в актуальное состояние
  * CLI для миграций https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
  * Пример смотри в db/migrations и скрипты migrate- up/down в Makefile
* Стиль кода https://github.com/uber-go/guide/blob/master/style.md