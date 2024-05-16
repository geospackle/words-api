# word api

### about
Stores alphabetic `[a-zA-Z]` words into and OpenSearch database and allows searching for top words by case-insensitive prefix.

### how to
 - start services: `docker-compose build && docker-compose up`
 - insert words to database: `curl -X POST -H "Content-Type:application/json" -d '{"word":"anyword"}' http://localhost:8080/words/add`
 - get top words by prefix: `curl http://localhost:8080/words/search\?prefix\="any"`

 NOTE: it may take a while for the database to spin up, retry in a minute on connection error.
