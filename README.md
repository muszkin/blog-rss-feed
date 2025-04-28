# blog-rss-feed

1. Dependencies:
   2. Postgress@15
   3. Go
4. Install gator app using:
   5. ```go install```
6. Create config file `~/.gatorconfig.json` with content:
   7. `{"db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable","current_user_name":""}`
8. Register new user using command `gator register [name]`
9. Log in using command `gator login [name]`
10. Add new feed `gator addfeed [name] [url]`
11. Start deamon as `gator agg 1h` to grab posts from rss feed every 1h
12. Use `gator browse` to check grabbed posts