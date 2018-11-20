grammar: 
	java org.antlr.v4.Tool -Dlanguage=Go -o parser Mdb.g4

compile:
	go build -v && go test && go vet

clean:
	rm -f *.class && rm -f *.interp && rm -f *.java && rm -f *.tokens

all: | grammar compile
	
