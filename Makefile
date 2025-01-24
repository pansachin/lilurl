# Build local image
build::
	@docker build  -t lilurl:latest -f Dockerfile .

# Run local image
run:: build
	@docker run -d -p 3000:3000 --name lilurl lilurl:latest


# Run test
test::
	@go test -v ./...


clean::
	@docker-compose down

rm::
	@docker rm -f lilurl


SERVICE := 'lilurl'
# Create tag for image
# Applicable to artifact registry images
ifdef GITHUB_RUN_NUMBER
	TAG = $(GITHUB_RUN_NUMBER)-$(GITHUB_REF_NAME)
endif

# Build image for artifact registry
build-ar::
	@docker build -t $(GAR)/$(SERVICE):$(TAG) -f Dockerfile .


# Push image to artifact registry
push-ar::
	@docker push $(GAR)/$(SERVICE):$(TAG)

migrate::
	@dbmate up 

