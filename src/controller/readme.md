## Dev steps
### Generate models from DB

To generate models structs from exsisting database it's used `xo` CLI tool [Source](https://github.com/xo/xo) 

```console
# Install tool
go get -u github.com/xo/xo

# Generate model from database
cd db
xo "pgsql://postgres:postgrespass@localhost/heta-ci?sslmode=disable" --schema public
```