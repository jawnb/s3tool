s3tool
=======

Fast and simple golang powered S3 management utility.

Created by John Begeman

Usage
-----

s3tool supports a number of commands, each supporting file globbing.

## Show whats in a bucket

    s3tool ls s3://your.bucket/cat-pictures.*
    
    Would show all your cat-pictures

## Remove stuff from a bucket

    s3tool rm s3://your.bucket/dog-pictures.*

Would remove all dog pictures from your.bucket 


## Downloading from a bucket
    s3tool get s3://your.bucket/cat-pictures.*

Will download all cat pictures from your.bucket to the current working directory


Build and Install
-----------------
    go get launchpad.net/goamz/aws
    go build s3tool.go
    export AWS_SECRET_ACCESS_KEY="YOUR-SUPER-SECRET-KEY"
    export AWS_ACCESS_KEY_ID="YOUR-ACCESS-ID"
    ./s3tool ls s3://path.to.your.bucket/*.jpg


