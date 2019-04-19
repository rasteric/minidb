compile:
	go build -v && go test && go vet

clean:
	rm -f *.class && rm -f *.interp && rm -f *.java && rm -f *.tokens

all: | compile
	
