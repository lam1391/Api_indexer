
# API INDEXER

This is a API sever and methods to connect the front with zinc Search

methods


## API Reference

#### Get all mails

```http
  GET /api//mails/?{from}&{max} 
```

| Parameter | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `from` | `int` | **Required**. from > 0 and <= total results|
| `max` | `int` | **Required**.  max > 0 and max <= 100 |

#### Get item

```http
  GET /api//mails/filter/?{from}&{max}&{filterID}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `from` | `int` | **Required**. from > 0 and <= total results |
| `max` | `int` | **Required**.  max > 0 and max <= 100 |
| `filterID` | `int` | **Required**.  |


## Environment Variables

To run this project, you will need to add the following environment variables to your .env file

`USER_ZINC`

`PASS_ZINC`

`HOST_ZINC`



# Authentication Server 

Validates the user and password and generates a token, the token is needed to make requests to the API.

## Environment Variables

To run this project, you will need to add the following environment variables to your .env file


`SECRET_KEY`

`PORT`



## Examples
Generate Token using username & password

```Curl
    POST http://localhost:3000/token
    User-Agent: Fiddler
    Host: localhost:3000
    Content-Length: 50
    Content-Type: application/x-www-form-urlencoded

    grant_type=password&username=user01&password=12345
}
```

