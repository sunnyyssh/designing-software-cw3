# designing-software-cw3

## Run

```shell
docker compose up
```

## Test

Imagine you are a user with ID `140bcaed-e10a-4fe8-bf7b-b829334f2d64`

```shell
export USER_ID=140bcaed-e10a-4fe8-bf7b-b829334f2d64
```

1. Make sure that such account doesn't exist

```shell
curl -X GET "localhost/payment/account/$USER_ID"
```

2. Create your account

```shell
curl -X PUT "localhost/payment/account/$USER_ID"
```

3. Put some money directly to your account

```shell
curl -X POST "localhost/payment/account/$USER_ID/amount" -d '{"amount": 1000}'
```

4. Create an order

```shell
curl -X POST "localhost/order/order" -d "{\"user_id\": \"$USER_ID\", \"amount\": 100}"
```

5. Check orders

```shell
curl -X GET "localhost/order/order/all"
```

6. Check that the order was applied

```shell
curl -X GET "localhost/payment/account/$USER_ID"
```

7. Try to create an order with too big amount of money

```shell
curl -X POST "localhost/order/order" -d "{\"user_id\": \"$USER_ID\", \"amount\": 100000}"
```

8. Check that last order is cancelled

```shell
curl -X GET "localhost/order/order/all"
```
