mainfile = main.go
objdir = ./$(obj)
objname = $(obj)
winobjname = $(obj).exe

mac:
	go build -o $(objdir)/$(objname) $(objdir)/$(mainfile)
	go install $(objdir)
	@echo "======Build mac done======"
win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(objdir)/$(winobjname) $(objdir)/$(mainfile)
	mkdir $(GOPATH)/bin/win
	cp $(objdir)/$(winobjname) $(GOPATH)/bin/win/$(winobjname)
	@echo "======Build win done======"
wincc:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o $(objdir)/$(winobjname) $(objdir)/$(mainfile)
	mkdir $(GOPATH)/bin/win
	cp $(objdir)/$(winobjname) $(GOPATH)/bin/win/$(winobjname)
	@echo "======Build wincc done======"
linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(objdir)/$(objname) $(objdir)/$(mainfile)
	mkdir $(GOPATH)/bin/linux
	cp $(objdir)/$(objname) $(GOPATH)/bin/linux/$(objname)
	@echo "======Build linux done======"
linuxcc:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc go build -o $(objdir)/$(objname) $(objdir)/$(mainfile)
	mkdir $(GOPATH)/bin/linux
	cp $(objdir)/$(objname) $(GOPATH)/bin/linux/$(objname)
	@echo "======Build linuxcc done======"
