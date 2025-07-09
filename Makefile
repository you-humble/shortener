
.PHONY: build, run, test

build:
	go build -o bin/shortener cmd/shortener/*.go

up-db:
	docker compose -f deployments/docker-compose.yaml up -d

down-db:
	docker compose -f deployments/docker-compose.yaml down

run: build
	export SECRET_KEY=pidar
	bin/shortener -d postgres://eblan:ne_eblan@localhost:5432/shortener?sslmode=disable

test: build
	shortenertest -test.v -test.run=^TestIteration1$$ -binary-path=bin/shortener
	shortenertest -test.v -test.run=^TestIteration2$$ -source-path=.
	shortenertest -test.v -test.run=^TestIteration3$$ -source-path=.
	shortenertest -test.v -test.run=^TestIteration4$$ -source-path=. -binary-path=bin/shortener -server-port=3000
	shortenertest -test.v -test.run=^TestIteration5$$ -source-path=. -binary-path=bin/shortener -server-port=8080
	shortenertest -test.v -test.run=^TestIteration6$$ -source-path=. -binary-path=bin/shortener

testbeta: build
	shortenertestbeta -test.v -test.run=^TestIteration1$$ -binary-path=bin/shortener
	shortenertestbeta -test.v -test.run=^TestIteration2$$ -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration3$$ -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration4$$ -source-path=. -binary-path=bin/shortener -server-port=8080
	shortenertestbeta -test.v -test.run=^TestIteration5$$ -source-path=. -binary-path=bin/shortener -server-port=8080
	shortenertestbeta -test.v -test.run=^TestIteration6$$ -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration7$$ -source-path=. -binary-path=bin/shortener
	shortenertestbeta -test.v -test.run=^TestIteration8$$ -binary-path=bin/shortener
	shortenertestbeta -test.v -test.run=^TestIteration9$$ -source-path=. -binary-path=bin/shortener -file-storage-path=tmp/short-url-db.json
	shortenertestbeta -test.v -test.run=^TestIteration10$$ -source-path=. -binary-path=bin/shortener -database-dsn=postgres://eblan:ne_eblan@localhost:5432/shortener?sslmode=disable
	shortenertestbeta -test.v -test.run=^TestIteration11$$ -source-path=. -binary-path=bin/shortener -database-dsn=postgres://eblan:ne_eblan@localhost:5432/shortener?sslmode=disable
	shortenertestbeta -test.v -test.run=^TestIteration12$$ -binary-path=bin/shortener -database-dsn=postgres://eblan:ne_eblan@localhost:5432/shortener?sslmode=disable
	shortenertestbeta -test.v -test.run=^TestIteration13$$ -binary-path=bin/shortener -database-dsn=postgres://eblan:ne_eblan@localhost:5432/shortener?sslmode=disable
	shortenertestbeta -test.v -test.run=^TestIteration14$$ -binary-path=bin/shortener -database-dsn=postgres://eblan:ne_eblan@localhost:5432/shortener?sslmode=disable