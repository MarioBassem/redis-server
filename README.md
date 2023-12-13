# Redis Server

This is a simple redis server. It is part of the coding challenges <https://codingchallenges.fyi/challenges/intro>

## Componenets

- Resp parser
- Redis server
  - the server deserializes incoming requests
  - it should handle the following requests:
    - `GET`: get a value
    - `SET`: set a value
    - `EXISTS`: check if a key exists
    - `DEL`: delete an entry
    - `INCR`: incmrenet a stored number by one
    - `DECR`: decrement a stored number by one
    - `LPUSH`: insert all values at the head of a list
    - `RPUSH`: insert all values at the tail of a list
    - `SAVE`: save the database to disk
    - `PING`: the server should respond with a pong message
- Redis cli
  - a cli to generate resp serialized requests

### Design

- Server
  - server listens for incoming connections on port 6379
  - the server should handle requests concurrently
  - each request should require a suitable lock to modify the database's state
