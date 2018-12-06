# CrossRegionRDSBacker
Utility to Create the Snapshots of RDS in a specific AWS Region and Copy the Snapshot to the Cross Region in AWS i.e. a Destination Region.
Also This utility will clean the old snapshots which are older than Retention days specified. 

The Code can be run as a Cron or as a Lambda Function. 

This Utility runs in two modes 
Mode 1 - "create" - It will create a snapshot of the current date for all RDS instances in source Region and do the clean up of old snapshots. 
Mode 2 - "copy" - It will copy all snapshots created for the same day to the another region. This should be run after the "create" mode

# Install

```go get github.com/ankitschopra/CrossRegionRDSBacker```

```go install github.com/ankitschopra/CrossRegionRDSBacker```



# Download < In case you dont have Go setup>
Download from Release section - https://github.com/ankitschopra/CrossRegionRDSBacker/releases/

# How to Use
```
$GOPATH/bin/CrossRegionRDSBacker --help
Usage of ./CrossRegionRDSBacker:
  -destRDSRegion string
    	Region of AWS where Snapshot of Source Region RDS to be copied
  -mode string
    	This Script Runs in two mode, Mode 1 - "create": Create the snapshot of all RDS in source Region and delete old snapshots beyond retention days.
    	                 Mode 2 - "copy": It will copy the current date snapshot to the destination region.
  -retentionDays int
    	No of days the snapshots to be retained in Both Source and destination Region. Defaults to 3 (default 3)
  -sourceRDSRegion string
    	Region of AWS where RDS instances are present
 ```
      
      
# Limitation with AWS
currently there is limit of 5 concurrent snapshot at a time for Cross Region copy. Its a hard limmit from AWS. We need to add the functionality here in the script to handle it for now.

