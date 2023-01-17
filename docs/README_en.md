# GoodFS

GoodFS is a highly focused distributed system for reading and writing file objects, which has excellent high
availability, ultimate consistency, and powerful horizontal scalability. It is particularly suitable for scenarios where
there are more reads than writes. In terms of object storage, it adopts multi-replica strategy and erasure code file
repair technology, supports Bucket classification management, and flexible fine-grained object configuration, making it
easy for users to manage and maintain stored objects. The data migration function of GoodFS includes both metadata and
object data. Metadata migration involves file meta-information in the file system, including file names, sizes, and
creation times. Object data migration includes the data content of the file itself. These two parts of data migration
are independent of each other, and GoodFS has strict data consistency guarantees, ensuring that data is not lost or
duplicated during migration. In addition, GoodFS also supports scheduled migration and manual migration, allowing users
to migrate data according to their own needs. GoodFS provides various deployment methods, including monolithic
deployment, pseudo-distributed deployment, and fully distributed deployment. Monolithic deployment is the simplest
deployment method, which deploys all services on one machine. This method is suitable for small-scale or test
environments. Pseudo-distributed deployment deploys metadata services and object data services on different machines,
but the interface service is still deployed on a single machine. This method is suitable for medium-scale environments.
Fully distributed deployment deploys all services on different machines, which is suitable for large-scale environments.
The GoodFS architecture includes interface services, metadata services, object data services, and ETCD as a coordination
center. The interface service is responsible for processing user requests, the metadata service is responsible for
maintaining file meta-information, the object data service is responsible for maintaining file data, and ETCD, as the
coordination center, is responsible for maintaining cluster status and information.