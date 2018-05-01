mainfile = main.go
objdir = ./$(obj)
winobjname = $(obj).exe

mac:
	go build -o $(objdir)/$(objname) $(objdir)/$(mainfile)
	go install $(objdir)
	@echo "======Build mac done======"
win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(objdir)/$(winobjname) $(objdir)/$(mainfile)
	cp $(objdir)/$(winobjname) $(GOPATH)/bin/$(winobjname)
	@echo "======Build win done======"
wincc:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o $(objdir)/$(winobjname) $(objdir)/$(mainfile)
	cp $(objdir)/$(winobjname) $(GOPATH)/bin/$(winobjname)
	@echo "======Build wincc done======"
linux:
	@echo "Error: Unsupport this Platform..."
