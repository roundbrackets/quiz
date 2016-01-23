Western Digital
===============

First Interview
----------------
Programming language may vary but question pertains to concurrency, order and deadlocks. Actually, I think I tookit there myself.

### You have 3 threads, each thread has three queries

For example

update table1, table2 set table1.counter = table1.ounter + table2.increment

How does mysql lock here? Can you make MySQL lock as few resources as possible (row lock, column lock), and if so how?
There are auto incrementing features of mysql...

    LCCK TABLE

Is LOCK TABLE a thing?

    Yes...

Anything else?

    COLUMN LOCKS...

### Is there anoher way?

    Well, you could have a single mysql connection instance throu which all requsts are funnlled. What if we don't we have a connection pool.

### Ok, you could have a singleton which holds the queries, but then all the threads have a reference to the insgleton.

    *fumble* Static class, with instance of itself, synscronized. *mumle*

### Ok, you've done this on a filestem to prevent multiple cron jobs running in the same time.

    I've faced this problem with a scenario with queing. In that instance I did stat on the script file to see if it was open,
    I opened the script file, if not I opned it and then ran the script and at the end I closed the file.
    I could not remeber how to check if the file was opne (i should have said lsof and grep, because it's stupid and it works)

### What aboit file locks?

    Yes, I know of them, but not the commands. How abpout we just make a pid file?

Perhaps this would have been part a satisfying answer
-----------------------------------------------------

    Yes, TABLE LOCKs are a thing, as are COLUMN and ROW LOCKs.

    Nothing that I talked about solves the problem of order. We'd need some other logic, say an auto incrmenting field that gives us a tranaction id, if we only read and mysql autoincrements maybe locking isn't required, then meybe we have an OK solution. Still you have to have a thing that hands out the IDs.

    Since each thread has three queries and the thread presumeably knows the order we can combine them to a transaction and get the transaction id.

    Then let's say we stuff into kafka and tell kafka it's an ordered list and be done with it.
