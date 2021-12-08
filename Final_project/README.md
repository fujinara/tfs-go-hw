## Бот для торговли на криптовалютной бирже [kraken-demo](https://demo-futures.kraken.com/futures/PI_XBTUSD)

### Запуск и управление

* Для запуска бота нужно скомпилировать файл main.go в папке bot командой `go build main.go`, а затем произвести запуск бинарника, задав ключи для работы с биржей и с телеграм ботом в виде переменных окружения. Например:
`KRAKEN_API_KEY=my_api_key KRAKEN_SECRET=my_private_key TELEGRAM_APITOKEN=my_telegram_apitoken ./main`

* Для отправки ioc заявки надо сделать POST запрос на эндпоинт `/sendorder` с заданными параметрами следующего вида:
`curl -v -X POST -H "Content-Type: application/json" --data '{"side" : "buy", "ticker" : "pi_xbtusd", "size" : 7, "price" : 55000}' 'localhost:5000/sendorder'`

* Для запуска алгоритма "take_profit/stop_loss" отправляется POST запрос на эндпоинт `/strategy` с заданными параметрами. Например:
`curl -v -X POST -H "Content-Type: application/json" --data '{"ticker" : "pi_xbtusd", "size" : 3, "percentage" : 0.08}' 'localhost:5000/strategy'`

ticker - название инструмента, side - направление сделки, size - размер сделки, price - цена заявки, percentage - изменение цены в процентах, при которой срабатывает "take_profit/stop_loss"