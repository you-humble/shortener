
.PHONY: build, run, test

build:
	go build -o bin/shortener cmd/shortener/*.go

run: build
	bin/shortener

test: build
	shortenertest -test.v -test.run=^TestIteration1$$ -binary-path=bin/shortener
	shortenertest -test.v -test.run=^TestIteration2$$ -source-path=.
	shortenertest -test.v -test.run=^TestIteration3$$ -source-path=.
	shortenertest -test.v -test.run=^TestIteration4$$ -source-path=. -binary-path=bin/shortener -server-port=3000
	shortenertest -test.v -test.run=^TestIteration5$$ -source-path=. -binary-path=bin/shortener -server-port=8080