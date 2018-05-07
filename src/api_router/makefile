mainfile = main.go
objdir = ./$(obj)
objname = $(obj)
winobjname = $(obj).exe

mac:
	go build -o $(objdir)/$(objname) $(objdir)/$(mainfile)
	test -d $(GOPATH)/bin/mac || mkdir -p $(GOPATH)/bin/mac
	cp $(objdir)/$(objname) $(GOPATH)/bin/mac/$(objname)
	@echo "======Build mac done======"
win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(objdir)/$(winobjname) $(objdir)/$(mainfile)
	test -d $(GOPATH)/bin/win || mkdir -p $(GOPATH)/bin/win
	cp $(objdir)/$(winobjname) $(GOPATH)/bin/win/$(winobjname)
	@echo "======Build win done======"
wincc:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o $(objdir)/$(winobjname) $(objdir)/$(mainfile)
	test -d $(GOPATH)/bin/win || mkdir -p $(GOPATH)/bin/win
	cp $(objdir)/$(winobjname) $(GOPATH)/bin/win/$(winobjname)
	@echo "======Build wincc done======"
linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(objdir)/$(objname) $(objdir)/$(mainfile)
	test -d $(GOPATH)/bin/linux/$(obj)_app || mkdir -p $(GOPATH)/bin/linux/$(obj)_app
	cp $(objdir)/$(objname) $(GOPATH)/bin/linux/$(obj)_app/$(objname)
	@echo "======Build linux done======"
linuxcc:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc go build -o $(objdir)/$(objname) $(objdir)/$(mainfile)
	test -d $(GOPATH)/bin/linux/$(obj)_app || mkdir -p $(GOPATH)/bin/linux/$(obj)_app
	cp $(objdir)/$(objname) $(GOPATH)/bin/linux/$(obj)_app/$(objname)
	@echo "======Build linuxcc done======"
