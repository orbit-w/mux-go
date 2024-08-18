Bench:
	go test -v -run=^$ -bench .

GenerateCoverageReport:
	go test -coverprofile=coverage.txt