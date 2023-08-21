Описание API(мог бы сделать и в Swagger, но для такого масштаба проще ручками)

Примеры ответов:
УСПЕШНЫЙ ЗАПРОС
{
    "done": true
}

НЕУСПЕШНЫЙ ЗАПРОС
{
    "done": false,
    "message": "Пользователь не найден"
}

http://127.0.0.1:8080/api/new - POST
Новый платёж
{
    "id": 4, - id человека
    "email": "d", - email
    "sum": 26.3, - сумма операции
    "currency": "GJR" - валюта операции(3-буквенная аббревиатура)
}

http://127.0.0.1:8080/api/confirm - POST
Подтверждение платежа
{
    "id": 1, - id платежа
    "status": 2, - новый статус платежа(2 или 3)
    "api_key": "test" - апи-ключ авторизации(можно поменять в базе)
}

http://127.0.0.1:8080/api/bill?id=3
Получение платежа по id

http://127.0.0.1:8080/api/user_profile?id=4
http://127.0.0.1:8080/api/user_profile?mail=d
Получение платежей для пользователя

http://127.0.0.1:8080/api/cancel
Отмена платежа
{
    "id":6 - номер платежа
}



