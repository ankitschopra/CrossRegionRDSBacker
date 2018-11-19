package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)


var sess = session.Must(session.NewSessionWithOptions(session.Options{
Profile: os.Getenv("AWS_PROFILE"),
}))


func checkAWSEnvVar() bool{
	if os.Getenv("AWS_PROFILE") == "" {
		return false
	}
	return true
}
func createSnapshotName( input string) string {
	current_time := time.Now().UTC()
	//fmt.Println("The Current time is ", current_time.Format("2006-01-02"))

	//Creating the Snapshot Identifier.
	var buffer bytes.Buffer
	buffer.WriteString("lambda-snapshot-")
	buffer.WriteString(input)
	buffer.WriteString("-")
	buffer.WriteString(string( current_time.Format("2006-01-02")))
	return buffer.String()
}

func create_db_snapshot(db_instance_identifier string, awsRegion string) bool{

	//DB Snapshot will be created with Prefix -  lambda-snapshot.

	var svc = rds.New(sess, &aws.Config{
		Region: aws.String(awsRegion),
	})

	db_snapshot_identifier := createSnapshotName(db_instance_identifier)

	fmt.Println(db_snapshot_identifier)
	input := &rds.CreateDBSnapshotInput{
		DBInstanceIdentifier: aws.String(db_instance_identifier),
		DBSnapshotIdentifier: aws.String(db_snapshot_identifier),
	}

	_, err := svc.CreateDBSnapshot(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBSnapshotAlreadyExistsFault:
				fmt.Println(rds.ErrCodeDBSnapshotAlreadyExistsFault, aerr.Error())
			case rds.ErrCodeInvalidDBInstanceStateFault:
				fmt.Println(rds.ErrCodeInvalidDBInstanceStateFault, aerr.Error())
			case rds.ErrCodeDBInstanceNotFoundFault:
				fmt.Println(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
			case rds.ErrCodeSnapshotQuotaExceededFault:
				fmt.Println(rds.ErrCodeSnapshotQuotaExceededFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return false
	}

	fmt.Println()
	return true

}

func copy_db_snapshot(dbSnapshotIdentifier string, sourceAwsRegion string ,destAwsRegion string) bool {

	svc := rds.New(sess, &aws.Config{
		Region: aws.String(destAwsRegion),
	})

	sourceDBsnapshotARN := get_db_snapshot_arn(sourceAwsRegion, dbSnapshotIdentifier)
	if sourceDBsnapshotARN == "nil" {
		return false
	}
	input := &rds.CopyDBSnapshotInput{
		CopyTags: aws.Bool(true),

		DestinationRegion: aws.String(destAwsRegion),
		TargetDBSnapshotIdentifier: aws.String(dbSnapshotIdentifier),

		SourceDBSnapshotIdentifier: aws.String(sourceDBsnapshotARN),
		SourceRegion: aws.String(sourceAwsRegion),
	}

	_, err := svc.CopyDBSnapshot(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBSnapshotAlreadyExistsFault:
				fmt.Println(rds.ErrCodeDBSnapshotAlreadyExistsFault, aerr.Error())
			case rds.ErrCodeDBSnapshotNotFoundFault:
				fmt.Println(rds.ErrCodeDBSnapshotNotFoundFault, aerr.Error())
			case rds.ErrCodeInvalidDBSnapshotStateFault:
				fmt.Println(rds.ErrCodeInvalidDBSnapshotStateFault, aerr.Error())
			case rds.ErrCodeSnapshotQuotaExceededFault:
				fmt.Println(rds.ErrCodeSnapshotQuotaExceededFault, aerr.Error())
			case rds.ErrCodeKMSKeyNotAccessibleFault:
				fmt.Println(rds.ErrCodeKMSKeyNotAccessibleFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return false
	}

	return true
}

func get_db_snapshot_arn(awsRegion string, dbSnapshotIdentifier string) string {


	svc := rds.New(sess, &aws.Config{
		Region: aws.String(awsRegion),
	})

	input := &rds.DescribeDBSnapshotsInput{
		DBSnapshotIdentifier: aws.String(dbSnapshotIdentifier),
		IncludePublic:        aws.Bool(false),
		IncludeShared:        aws.Bool(true),
		SnapshotType:         aws.String("manual"),
	}

	result, err := svc.DescribeDBSnapshots(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBSnapshotNotFoundFault:
				fmt.Println(rds.ErrCodeDBSnapshotNotFoundFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return "nil"
	}
	//fmt.Println(result)
	return *result.DBSnapshots[0].DBSnapshotArn
}

func get_all_db_list(awsRegion string) []string {
	var list_of_all_db []string

	svc := rds.New(sess, &aws.Config{
		Region: aws.String(awsRegion),
	})

	input := &rds.DescribeDBInstancesInput{
	}

	for {

		result, err := svc.DescribeDBInstances(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case rds.ErrCodeDBInstanceNotFoundFault:
					fmt.Println(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return list_of_all_db
		}

		for _, dbInstance := range result.DBInstances {
			//fmt.Println(*dbInstance.DBInstanceIdentifier)
			list_of_all_db = append(list_of_all_db, *dbInstance.DBInstanceIdentifier)
		}
		if result.Marker == nil {
			break
		}
		input = &rds.DescribeDBInstancesInput{
			Marker: result.Marker,
		}
	}

	//fmt.Println(list_of_all_db,len(list_of_all_db),cap(list_of_all_db))
	return list_of_all_db
}

func remove_db_snapshot(awsRegion string, snapShotIdentifier string) bool {

	svc := rds.New(sess, &aws.Config{
		Region: aws.String(awsRegion),
	})

	input := &rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: aws.String(snapShotIdentifier),
	}

	result, err := svc.DeleteDBSnapshot(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeInvalidDBSnapshotStateFault:
				fmt.Println(rds.ErrCodeInvalidDBSnapshotStateFault, aerr.Error())
			case rds.ErrCodeDBSnapshotNotFoundFault:
				fmt.Println(rds.ErrCodeDBSnapshotNotFoundFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return false
	}

	fmt.Println(result)
	return true
}

func get_list_of_snapshots(awsRegion string, all_db_list []string) []string {
	var list_of_snapshots []string
	svc := rds.New(sess, &aws.Config{
		Region: aws.String(awsRegion),
	})

	// Since lots of Manual Snapshots are there and Filter option is not availanble in describe snapshots,
	// We will Describe the Manual Snapshots based on the DB instance Identifier and will delete the Snapshots which are
	// Older


	for _,dbIdentifier := range all_db_list {

		input := &rds.DescribeDBSnapshotsInput{
			DBInstanceIdentifier: aws.String(dbIdentifier),
			IncludePublic:        aws.Bool(false),
			IncludeShared:        aws.Bool(false),
			SnapshotType:         aws.String("manual"),
		}

		// Marker/Pagination  Support to be Added.

		result, err := svc.DescribeDBSnapshots(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case rds.ErrCodeDBSnapshotNotFoundFault:
					fmt.Println(rds.ErrCodeDBSnapshotNotFoundFault, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			continue
		}

		for _,snapshot := range result.DBSnapshots {
			list_of_snapshots = append(list_of_snapshots,*snapshot.DBSnapshotIdentifier)
		}
	}

	return list_of_snapshots

}

func Date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func filter_snapshots(listOfSnapshots []string, prefix string, retentiondays int64) []string {
	var filteredList []string
	for _,snapshot := range listOfSnapshots {


		if strings.HasPrefix(snapshot,prefix) {

			if len(snapshot) < 10 {
				fmt.Printf("Skiping this snapshot: %s as it its length is too less" , snapshot)
				continue
			}
			snapshot_date := snapshot[len(snapshot)-10:]
			datelist := strings.Split(snapshot_date, "-")
			//fmt.Println(datelist)
			if len(datelist) != 3 {
				fmt.Printf("Skiping this snapshot: %s as its not having valid date in end", snapshot)
				continue
			}
			year,err1 := strconv.Atoi(datelist[0])
			mon,err2 := strconv.Atoi(datelist[1])
			day,err3 := strconv.Atoi(datelist[2])

			if err1 != nil || err2 != nil || err3 != nil {
				fmt.Printf("Skiping this snapshot: %s as its not having valid date in end", snapshot)
				continue
			}
			//fmt.Println(year,mon,day)
			snapshot_time := Date(year,mon,day)
			current_time := time.Now().UTC()
			days := current_time.Sub(snapshot_time).Hours() / 24
			//fmt.Printf("Snapshot %s is of %f Age. ", snapshot, days )
			if int64(days) >= retentiondays {
				filteredList = append(filteredList,snapshot)
			}
		}
	}
	//fmt.Println(filteredList)
	return filteredList
}



func createAllRdsSnapshot(awsRegionSource string, awsRegionDest string, retentionDays int64) {

	// Lets get the list of All DBs

	db_list := get_all_db_list(awsRegionSource)

	// Creating DB snapshot of all DB Instances.
	for _,db := range db_list {
		if !(create_db_snapshot(db,awsRegionSource )) {
			fmt.Println("DB Snapshot Creation Failed for DB: ",db)
		} else {
			fmt.Println("DB Snapshot Creation Succeeded for DB: ",db)
		}

	}


	// Deleting Old Snapshots from Both Source and Destination Region

	// Deleting from source Region

	list_of_snapshots := get_list_of_snapshots(awsRegionSource, db_list)
	filtered_list_of_snapshots :=  filter_snapshots(list_of_snapshots,"lambda-snapshot", retentionDays)
	for _,snapshot := range filtered_list_of_snapshots {

		if !(remove_db_snapshot(awsRegionSource, snapshot)) {
			fmt.Printf("DB Snapshot Deleting Failed for Snapshot: %s in Region: %s ", snapshot, awsRegionSource )
		} else {
			fmt.Printf("DB Snapshot Deleting Scucceeded for Snapshot: %s in Region: %s ", snapshot, awsRegionSource )
		}

	}

	// Deleting from Dest Region

	list_of_snapshots = get_list_of_snapshots(awsRegionDest, db_list)
	filtered_list_of_snapshots =  filter_snapshots(list_of_snapshots,"lambda-snapshot", retentionDays)
	for _,snapshot := range filtered_list_of_snapshots {

		if !(remove_db_snapshot(awsRegionDest, snapshot)) {
			fmt.Printf("DB Snapshot Deleting Failed for Snapshot: %s in Region: %s ", snapshot, awsRegionDest )
		} else {
			fmt.Printf("DB Snapshot Deleting Scucceeded for Snapshot: %s in Region: %s ", snapshot, awsRegionDest )
		}

	}

}

func copyCrossRegionAllSnapshot(awsRegionSource string, awsRegionDest string) {

	dbList := get_all_db_list(awsRegionSource)

	// Copying DB snapshot to Destination Region.
	for _,db := range dbList {
		dbSnapshotIdentifier := createSnapshotName(db)

		if !(copy_db_snapshot(dbSnapshotIdentifier,awsRegionSource,awsRegionDest)) {
			fmt.Printf("DB Snapshot Copy to Cross Region Failed for Snapshot: %s to Dest Region: %s ", dbSnapshotIdentifier, awsRegionDest )
		} else {
			fmt.Printf("DB Snapshot Copy to Cross Region Succeeded for Snapshot: %s to Dest Region: %s ", dbSnapshotIdentifier, awsRegionDest )
		}
	}

}

func main() {
	//cred_resposne, err := sess.Config.Credentials.Get()
	//fmt.Println("Cred Resposne", cred_resposne)
	//fmt.Println(err)


	sourceRDSRegion := flag.String("sourceRDSRegion", "", "Region of AWS where RDS instances are present")
	destRDSRegion := flag.String("destRDSRegion", "", "Region of AWS where Snapshot of Source Region RDS to be copied")
	retentionDays := flag.Int64("retentionDays", 3, "No of days the snapshots to be retained in Both Source and destination Region. Defaults to 3")
	mode := flag.String("mode", "", `This Script Runs in two mode, Mode 1 - "create": Create the snapshot of all RDS in source Region and delete old snapshots beyond retention days.
                 Mode 2 - "copy": It will copy the current date snapshot to the destination region.`)

	flag.Parse()

	// make sure destFile is defined in all cases
	if *sourceRDSRegion == "" || *destRDSRegion == ""  {
		fmt.Println("Both RDS source Region and Destination Region are mandatory.")
		fmt.Println(`Please provide "-sourceRDSRegion" and  "-destRDSRegion" options`)
		fmt.Println("This utility does not print out data on STDOUT")
		os.Exit(1)
	}

	if *mode == "" {
		fmt.Println("mode Paramater should be passed")
		fmt.Println(`Please provide "-mode" parameter with option either "copy" or "create" `)
		fmt.Println("This utility does not print out data on STDOUT")
		os.Exit(1)
	}


	if *mode != "copy" &&  *mode != "create" {
		fmt.Println(`mode parameter should have value - "copy" or "create"`)
	}

	// check if AWS credentials are present as environment variables
	if !checkAWSEnvVar() {
		log.Fatal(`AWS environment variables "AWS_PROFILE" is not defiled`)
		os.Exit(1)
	}

	if *mode == "copy" {
		copyCrossRegionAllSnapshot(*sourceRDSRegion, *destRDSRegion)
	}

	if *mode == "create" {
		createAllRdsSnapshot(*sourceRDSRegion, *destRDSRegion, *retentionDays)
	}

	//createAllRdsSnapshot("eu-west-1", "ap-southeast-1",2)

}




