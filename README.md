# employee-satisfaction

System for registering employee satisfaction around your office / offices.
Results will be saved in a Redshift datawarehouse.

See below for the overall architecture

![Architecture](/docs/employee-satisfaction-flow.png?raw=true "Architecture")

## Building & Deploying

### Requirements

You will need to have golang installed and configured correctly for everything to work.

[golang.org](https://golang.org/dl/)

You will need an S3 bucket to store the code on while deploying. This can be a general bucket and the code will be stored under the prefix `lambdas/` on that bucket.

But it's important that the ACCESSKEY+SECRETKEY / PROFILE you deploy with can write and read from this S3 bucket.

### Config

The following ENV vars should be set before building this program.

```text
OWNER = The owner of the stack, can be person, team etc.
ENVIRONMENT = The environment the stack should belong to (dev, test, prod etc). Defaults to dev.
S3BUCKET = The S3 bucket to store the code on when deploying.
APPLICATION_NAME = Name of the application. Defaults to employee-satisfaction
```

Note that you will need to run the `make persistent` target first and update the `cloudformation/template.yaml` file with the correct password for your redshift server.

The password should be KMS encrypted and the KMS key is created when you run the `make persistent` target.

Edit the `Mappings` in the `template.yaml` file to correspond with your parameters, redshift servers, email etc for the different Evironments.

### Encrypt parameters with KMS

### Building

`make deploy-persistent`

Encrypt the password for your redshift cluster using the provided KMS key and enter that and other parameters
in the `cloudformation/template.yaml` Mappings. The password should be encrypted hash base64 encoded.

`make deploy-serverless`

## Adding IoT Buttons

Just use the `AWS BTN Dev` mobile app to add your buttons to your account.

Make sure the region is the correct one and that you choose the `employee-satisfaction-publishEvent-(ENV)` lambda function as function.

## Usage

### Using the button

Pressing the button once (or twice quickly) will register as "good" and a long press (>1 second) will register as "bad".
Please see docs under `docs/` in the repository.
