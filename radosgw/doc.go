// Package radosgw implements the facilities to manage the radosgw service by admin OP
// API.  It supports usage, user, bucket, quota, capability management. Use the following
// code snippet to operate the radosgw service:
//     client, _ := radosgw.NewClient({endpoint}, {ak}, {sk})
//     client.GetUser({uid})
//     client.CreateUser({UserModel})
//     ...
// One can use any other APIs that supported by the package:
//  - user management
//    GetUser(uid ...string) (int, *UserType, error)
//    CreateUser(uid, dispName, email string, maxBuckets int) (int, *UserType, error)
//    UpdateUser(uid, displayName, email string,
//        maxBuckets int, suspended bool) (int, *UserType, error)
//    DeleteUser(uid string, purgeData bool) (int, error)
//    CreateKey(uid string) (int, []KeyType, error)
//    DeleteKey(uid, ak string) (int, error)
//    AddCaps(uid string, user, buckets, usage []string) (int, []CapType, error)
//    DeleteCaps(uid string, user, buckets, usage []string) (int, error)
//    GetQuota(uid, quotaType string) (int, *QuotaType, error)
//    SetQuota(uid, quotaType string, maxObjects,
//        maxSize int64, enabled bool, bucketName string) (int, error)
//  - bucket management
//    GetBucket(bucket, uid string, stats bool) (int, []BucketType, error)
//    DeleteBucket(bucket string, purgeObjects bool) (int, error)
//    DeleteObject(bucket, object string) (int, error)
//    LinkBucket(bucket, bucketId, uid string) (int, error)
//    UnlinkBucket(bucket, uid string) (int, error)
//    CreateBucket(bucket, region string) (int, error)
//  - usage management
//    GetUsage(uid string, start, end *time.Time,
//        showSummary, showEntries bool) (int, *UsageType, error)
//    DeleteUsage(uid string, start, end *time.Time, deleteAll bool) (int, error)
//
// All admin OP API performs the http request to the given radosgw service using the AWS
// S3(v4) signature method. The status code and raw bytes body of http response are all
// directly returned to the caller allowing you to define custom post-process strategies.
package radosgw
