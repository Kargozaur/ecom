.PHONY: gateway-docker user-docker

include proto/user.makefile
include svc/gateway/gateway.makefile
include svc/reg/user.makefile
include migrations/user/migrate.makefile

gateway-docker:
	docker build -f svc/gateway/dockerfile -t gateway .

user-docker:
	docker build -f svc/reg/dockerfile -t user .
