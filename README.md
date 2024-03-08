# pbandj
Read-then-burn esque pastebin. Each entry can only be viewed once.

## Running
### Build local
```sh
make build
./pbandj -port 1337
```
### Docker 
```sh
docker build .
docker run -p 1337:1337 --name pbandj <image_id>
```

## UI
A basic form is rendered at `/{id}`. This can be used to create a new paste, or will display an existing one if `{id}` is present.


## API
```
POST /paste
application/json
{
  content  string
}
{
  id      string
}
```  

```
GET /paste/{id}
application/json
{
  id      string
  content string
}
```
