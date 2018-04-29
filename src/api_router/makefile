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
linux:
    @echo "Error: Unsupport this Platform..."
