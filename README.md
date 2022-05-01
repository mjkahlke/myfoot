# myfoot
A command line utility to display your MySQL database footprint.

# Overview
**myfoot** compares the size of MySQL innodb database tables to their corresponding ibd files on the filesystem and generates some basic statistics. Efficiency and overhead percentages give a rough idea of the extent to which your database tables are fragmented. Whole volumes have been written on this subject and I'm certainly not going to get into database fragmentation, B-trees and nodes, buffer pool sizes, etc. But generally, one can expect databases to become more fragmented, the ratio of data and meta-data to storage will decrease, and performance may drop as tables become more fragmented.

# How it works
**myfoot** takes no parameters. It just requires two environment variables to be set:
* `DBUSER` - MySQL account user
* `DBPASS` - MySQL account password

These credentials are used to open a connection to the local MySQL service on TCP port 3306, then query the _information_schema_ database for information about your databases and tables therein. For each schema and table, the associated _.ibd_ files are checked to determine the filesystem footprint, then generates efficiency and overhead percentages. Efficiency is the ratio of space taken up by data compared to space taken up on the filesystem. Overhead is the percentage of index or metadata to the sum of index, data, and free space taken up by the buckets reserved for the table.

# Epilogue
I used this project as a quick and easy way to learn Go. It's based on a Perl script I wrote years ago that pretty much did the same thing.
