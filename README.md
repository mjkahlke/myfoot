# myfoot
A command line utility to display your MySQL database footprint.

## Overview
**myfoot** compares the size of MySQL innodb database tables to their corresponding _.ibd_ files on the filesystem to generate efficiency and overhead percentages. These basic statistics give a rough idea of the extent to which your database tables are fragmented. Whole volumes have been written on database fragmentation, B-trees and nodes, buffer pool sizes, etc., none of which we get into here. But generally, as data is inserted and updated,  the ratio of data and meta-data to storage will decrease, and performance may drop as tables become more fragmented.

## How it works
**myfoot** takes no parameters. It just requires two environment variables to be set:
* `DBUSER` - MySQL account user
* `DBPASS` - MySQL account password

These credentials are used to open a connection to the local MySQL service on TCP port 3306 and query the _information_schema_ database for sizing information about your tables. For each schema and table, the associated _.ibd_ files are checked to determine the filesystem footprint. From this data, efficiency and overhead percentages are determined. _Efficiency_ is the ratio of space taken up by data compared to space taken up on the filesystem. _Overhead_ is the percentage of index to the sum of index, data, and free space taken up by the buckets reserved for the table (basically the ratio of metadata to data).

## Epilogue
I used this project as a quick and easy way to learn how to access databases in Go. It's based on a Perl script I wrote years ago that pretty much did the same thing.
