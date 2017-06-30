all:
	cd logProcess;go build -o logProcess *.go
	cd multi_handler_sample;go build -o logProcess *.go
clean:
	rm -fv logProcess/logProcess multi_handler_sample/logProcess
