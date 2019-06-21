# Project variables
IMAGE_NAME ?= azzart/${CIRCLE_PROJECT_REPONAME}
#VERSION ?= $(shell git describe --abbrev=0 --tags)
#VERSION ?= "$(shell semantics --output-tag)"
#            tag=$(semantics --output-tag)
#            if [ "$tag" ]; then
#              ghr -t $GITHUB_TOKEN -u $CIRCLE_PROJECT_USERNAME -r $CIRCLE_PROJECT_REPONAME --replace $tag dist/
#            else
#              echo "The commit message(s) did not indicate a major/minor/patch version."
#            fi


.PHONY: all build push help

.DEFAULT_GOAL := default

default: help ;

all: build test push

TAG ?= $(shell git describe --abbrev=0 --tags)
TAG_SHORT ?= $(shell git describe --abbrev=0 --tags | tr -d v)

create_release_tag:
	${INFO} "Creating release tag..."
	tag=$(shell semantics --output-tag); \
	[ "$${tag}" ] && echo "all good. tag is: $${tag}" || ( echo "tag is not set"; exit 1 )

build_bins:
	${INFO} "Building bins... $(IMAGE_NAME):${TAG}"
	@ gox -os="linux darwin windows" -arch="amd64" -output="dist/aws_es_proxy_{{.OS}}_{{.Arch}}" ./cmd/server

build_image:
	${INFO} "Building image... $(IMAGE_NAME):${TAG_SHORT}"
	@ docker build -t $(IMAGE_NAME):${TAG_SHORT} --no-cache .

push_github:
	${INFO} "Pushing bins to GitHub... $(IMAGE_NAME):${TAG}"
	@ ghr -t ${GITHUB_TOKEN} -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace ${TAG} dist/

push_dockerhub: login tag_latest
	${INFO} "Publishing docker image to DockerHub... $(IMAGE_NAME):${TAG_SHORT}"
	@docker push $(IMAGE_NAME):${TAG_SHORT}
	@docker push $(IMAGE_NAME):latest
	${INFO} "Publish complete"

tag_latest:
	${INFO} "Tagging image... $(IMAGE_NAME):${TAG_SHORT} as latest"
	@docker tag $(IMAGE_NAME):${TAG_SHORT} $(IMAGE_NAME):latest

login:
	${INFO} "Logging in to DockerHub..."
	@ echo ${DOCKER_PWD} | docker login -u ${DOCKER_LOGIN} --password-stdin
	${INFO} "Logged in to DockerHub"

help:
	${INFO} "-----------------------------------------------------------------------"
	${INFO} "                      Available commands                              -"
	${INFO} "-----------------------------------------------------------------------"
	${INFO} "   > build - To build $(CURRENT_DIR) image."
	${INFO} "   > push - To push $(CURRENT_DIR) image."
	${INFO} "   > all - To execute all steps."
	${INFO} "   > help - To see this help."
	${INFO} "-----------------------------------------------------------------------"

# Cosmetics
RED := "\e[1;31m"
YELLOW := "\e[1;33m"
NC := "\e[0m"

# Shell Functions
INFO := @bash -c '\
  printf $(YELLOW); \
  echo "=> $$1"; \
  printf $(NC)' SOME_VALUE
