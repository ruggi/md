MAIN_BRANCH = main

GOLANGCI_VERSION = 1.32.0
GOLANGCI = .bin/golangci/$(GOLANGCI_VERSION)/golangci-lint

CURRENT_VERSION_MAJOR = 1
CURRENT_VERSION_MINOR = 9
CURRENT_VERSION_BUG = 0

$(GOLANGCI):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(dir $(GOLANGCI)) v$(GOLANGCI_VERSION)

.PHONY: lint
lint: $(GOLANGCI)
	$(GOLANGCI) run ./...

.PHONY: test
test:
	go test ./...

publish:
	@if [ "$(VERSION)" = "" ] ; then echo You should define the version like so: make publish VERSION=x.y.z ; exit 1 ; fi
	@git diff --exit-code --cached || { git status ; echo You have changes that are staged but not committed ; false ; };
	@git diff --exit-code || { git status ; echo You have changes that are not committed ; false ; };
	@git diff --exit-code Makefile || { echo You have made changes to the Makefile that were not committed, please stash or commit them ; false ; };
	$(eval dots = $(subst ., ,$(VERSION)))
	$(eval new_major = $(word 1, $(dots)))
	$(eval new_minor = $(word 2, $(dots)))
	$(eval new_bug = $(word 3, $(dots)))
	sed -i.bak -e 's/^\(var Version = \).*/\1"$(VERSION)"/g' version/version.go
	sed -i.bak -e 's/^\(CURRENT_VERSION_MAJOR = \).*/\1$(new_major)/g' Makefile
	sed -i.bak -e 's/^\(CURRENT_VERSION_MINOR = \).*/\1$(new_minor)/g' Makefile
	sed -i.bak -e 's/^\(CURRENT_VERSION_BUG = \).*/\1$(new_bug)/g' Makefile
	rm Makefile.bak version/version.go.bak

	git commit -am 'Bump version to v$(VERSION)'
	git tag v$(VERSION)
	git push --follow-tags
	git push origin v$(VERSION)

update-main:
	git checkout $(MAIN_BRANCH)
	git pull

publish-major: update-main
	@make publish VERSION=$$(($(CURRENT_VERSION_MAJOR) + 1)).0.0
publish-minor: update-main
	@make publish VERSION=$(CURRENT_VERSION_MAJOR).$$(($(CURRENT_VERSION_MINOR) + 1)).0
publish-patch: update-main
	@make publish VERSION=$(CURRENT_VERSION_MAJOR).$(CURRENT_VERSION_MINOR).$$(($(CURRENT_VERSION_BUG) + 1))
