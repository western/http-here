
TOPDIR=$(PWD)
ASSETS=$(TOPDIR)/internal/app/assets

tinymce760:
	cd $(ASSETS) && \
	wget https://download.tiny.cloud/tinymce/community/tinymce_7.6.0.zip && unzip tinymce_7.6.0.zip && \
	mv tinymce tinymce_7.6.0 && rm tinymce_7.6.0.zip ; \

bootstrap533:
	cd $(ASSETS) && \
	wget https://github.com/twbs/bootstrap/releases/download/v5.3.3/bootstrap-5.3.3-dist.zip && unzip bootstrap-5.3.3-dist.zip && \
	mv bootstrap-5.3.3-dist bootstrap-5.3.3 && rm bootstrap-5.3.3-dist.zip ; \

bicons1131:
	cd $(ASSETS) && \
	wget https://github.com/twbs/icons/releases/download/v1.13.1/bootstrap-icons-1.13.1.zip && unzip bootstrap-icons-1.13.1.zip && \
	rm bootstrap-icons-1.13.1.zip ; \

without_cdn_build: tinymce760 bootstrap533 bicons1131
	cd $(TOPDIR)/internal/app/view/layout && \
	mv default.html default.html_ && mv default.nocdn.html default.html && \
	mv error.html error.html_ && mv error.nocdn.html error.html && \
	cd $(TOPDIR)/internal/app/view/edit && \
	mv edit_doc.html edit_doc.html_ && mv edit_doc.nocdn.html edit_doc.html && \
	cd $(TOPDIR) && make build

without_cdn_clear:
	rm http-here && \
	rm internal/app/view/layout/default.html_ && \
	rm internal/app/view/layout/error.html_ && \
	\
	cd $(ASSETS) && \
	rm -rf bootstrap-5.3.3 && \
	rm -rf bootstrap-icons-1.13.1 && \
	rm -rf tinymce_7.6.0 && \
	\
	git restore internal/app/view/layout

fmt:
	go fmt internal/api/* && go fmt internal/app/*.go && go fmt internal/cert/* && go fmt internal/model/* && go fmt internal/util/*

tidy:
	go mod tidy

run:
	go run cmd/http-here/main.go

build:
	go build -o http-here -ldflags "-w -s" cmd/http-here/main.go


