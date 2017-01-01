all:
	cd logProcess;go build -o logProcess *.go
clean:
	rm -fv logProcess/logProcess
