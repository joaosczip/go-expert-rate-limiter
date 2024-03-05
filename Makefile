test:
	siege -d0 -c20 -r1 http://localhost:8080

test-api-key:
	siege -d0 -c20 -r1 --header="API_KEY: 123" http://localhost:8080