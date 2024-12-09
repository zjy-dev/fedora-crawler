RPM_URL_FILE_FULLPATH_X64 := ./fedora_rpm_url_x64.txt
RPM_URL_FILE_FULLPATH_AARCH64 := ./fedora_rpm_url_aarch64.txt

fedora-crawler: init
	@echo "Building fedora-crawler"
	@go build -o ../fedora-crawler .

crawl-fedora: init fedora-crawler 
	@echo "Crawling fedora packages url"
	@./fedora-crawler $(RPM_URL_FILE_FULLPATH_X64) $(RPM_URL_FILE_FULLPATH_AARCH64)

init:
	@go mod download -x
	@touch $(RPM_URL_FILE_FULLPATH_X64) $(RPM_URL_FILE_FULLPATH_AARCH64)
