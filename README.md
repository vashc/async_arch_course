Awesome Task Exchange System
===

## Запуск
Для запуска можно воспользоваться
```makefile
make run
```

Для остановки работы
```makefile
make stop
```

В директории tests лежит коллекция с ручками для Postman.

### Миграция с разделением поля tasks.title
- добавляем новое поле tasks.jira_id в БД миграцией со значением по умолчанию (обратная совместимость)
- создаём новую схему v2 с добавлением поля jira_id и переводим на неё консьюмеры (обратная совместимость)
- переводим на схему v2 продьюсеры

### Обработка ошибок
События, которые по какой-то причине не смогли обработаться консьюмерами, могут быть отправлены
в DLQ очереди (например, через соответствующий exchange). В таких очередях можно выставить TTL,
по истечению которого сообщения отправятся заново в овносной exchange для обработки.

Для сообщений в DLQ очередях можно установить лимит повторов обработки, при превышении которого
можно отправлять сообщения в специальную очередь-отстойник без консьюмеров, в которой можно проводить
разбор причин ошибок (например, была использована некорректная версия схемы события).

Кроме того, можно настроить алерты, если количество сообщений в DLQ превышает определённый лимит, чтобы
можно было оперативно среагировать.

<details open>
   <summary><strong>Составляющие требований</strong></summary>

   ### Auth
   В текущем варианте данные о пользователях хранятся в каждом сервисе, создаются по событию user_created.

   ![](docs/auth_components.png)

   ### Task
   ![](docs/task_components.png)

   ### Account
   ![](docs/account_components.png)
   
   ### Analytics
   ![](docs/analytics_components.png)

</details>

<details open>
   <summary><strong>Модель данных</strong></summary>

   ![](docs/data_model.png)

</details>

<details open>
   <summary><strong>Доменная модель</strong></summary>

   ![](docs/domain_model.png)

</details>

<details open>
   <summary><strong>Сервисы</strong></summary>

   ![](docs/services.png)

</details>

<details open>
   <summary><strong>BC/CUD события</strong></summary>

   ![](docs/bc_cud_events.png)

</details>
