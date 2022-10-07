## Temporary targets to test sync APIs
create-relationship:
	./bin/galadriel-server create member -t td1
	./bin/galadriel-server create member -t td2
	./bin/galadriel-server create relationship -a td1 -b td2

.PHONY: create-relationship

run-harvester: create-relationship
	token=`./bin/galadriel-server generate token -t td2 | grep -Po "(?<=Access Token:\s).*"`; \
	./bin/galadriel-harvester run -t $$token

test-sync:
	token=`./bin/galadriel-server generate token -t td1 | grep -Po "(?<=Access Token:\s).*"`; \
	curl 127.0.0.1:8085/bundle/sync \
		-X "POST" \
		-H "Authorization: Bearer $$token" \
		-d "@dev/request_data/bundle_sync.json" | jq

test-post:
	token=`./bin/galadriel-server generate token -t td2 | grep -Po "(?<=Access Token:\s).*"`; \
	curl 127.0.0.1:8085/bundle \
		-X "POST" \
		-H "Authorization: Bearer $$token" \
		-d "@dev/request_data/bundle_post.json" | jq
