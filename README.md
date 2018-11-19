# CrossRegionRDSBacker
Utility to Create the Snapshots of RDS in a specific AWS Region and Copy the Snapshot to the Cross Region in AWS i.e. a Destination Region.
Also This utility will clean the old snapshots which are older than Retention days specified. 

The Code can be run as a Cron or as a Lambda Function. 

This Utility runs in two modes 
Mode 1 - "create" - It will create a snapshot of the current date for all RDS instances in source Region and do the clean up of old snapshots. 
Mode 2 - "copy" - It will copy all snapshots created for the same day to the another region. This should be run after the "create" mode

