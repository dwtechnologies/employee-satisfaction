# General
APPLICATION_NAME ?= employee-satisfaction
ENVIRONMENT      ?= dev
S3BUCKET         ?= our.code.s3.bucket

# Tags
OWNER ?= OwnerOfProject

.PHONY: build-go deploy-persistent package-serverless deploy-serverless
build-go:
	./build-go.sh

.PHONY: deploy-persistent
deploy-persistent:
	aws cloudformation deploy \
		--template-file cloudformation/persistent.yaml \
		--stack-name $(APPLICATION_NAME)-persistent-$(ENVIRONMENT) \
		--capabilities CAPABILITY_NAMED_IAM \
		--no-fail-on-empty-changeset \
		--parameter-overrides \
			ApplicationName=$(APPLICATION_NAME) \
			Env=$(ENVIRONMENT) \
		--tags \
			Project=$(APPLICATION_NAME) \
			Environment=$(ENVIRONMENT) \
			Owner=$(OWNER)

package-serverless:
	aws cloudformation package \
		--template-file cloudformation/template.yaml \
		--output-template-file dist/template-$(ENVIRONMENT).yaml \
		--s3-bucket $(S3BUCKET) \
		--s3-prefix lambdas/$(APPLICATION_NAME)-$(ENVIRONMENT)

deploy-serverless: build-go package-serverless
	aws cloudformation deploy \
		--template-file dist/template-$(ENVIRONMENT).yaml \
		--stack-name $(APPLICATION_NAME)-serverless-$(ENVIRONMENT) \
		--capabilities CAPABILITY_NAMED_IAM \
		--no-fail-on-empty-changeset \
		--parameter-overrides \
			ApplicationName=$(APPLICATION_NAME) \
			Env=$(ENVIRONMENT) \
		--tags \
			Project=$(APPLICATION_NAME) \
			Environment=$(ENVIRONMENT) \
			Owner=$(OWNER)


.PHONY: delete delete-persistent delete-serverless clean
delete: delete-serverless delete-persistent

delete-persistent:
	aws cloudformation delete-stack \
		--stack-name $(APPLICATION_NAME)-persistent-$(ENVIRONMENT)

delete-serverless:
	aws cloudformation delete-stack \
		--stack-name $(APPLICATION_NAME)-serverless-$(ENVIRONMENT)

clean:
	rm -r dist/*