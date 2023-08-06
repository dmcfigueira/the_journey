# Context
This repository contains (an incomplete) solution to the Staking Rewards coding challenge - [The Journey](https://github.com/stakingrewards/engineering-challenge/tree/backend).

The solution is implemented in Golang - a language which I never used before and picked up specificaly for this task. Nevertheless, Golang caught my attention
for its simplicity to use - resembling somewhat of a more modern C language (I'm usually a C/C++ programmer). 

Besides it's worry-free memory management, Go seems to be designed for multi-core processing, promising an easy implementation of concurrency with the
use of _goroutines_ - a not so easy task on some other languages. As such I decided to design my application in such a way that I could take advantage of this, and use _goroutines_ to solve the problem concurrently.

In the end you can argue it was over-engineered (which is not false) - but I'll say it was part of the journey :wink:.

# Usage
Simply install [Go (version 1.2+)](https://go.dev/doc/install), clone the repository to a new folder and open a terminal at that location.
To compile the program use:

> go build

This should generate an executable*. To run the program (on Windows) use: 

> .\the_journey.exe transactions.csv

<sub>*note: prior to building the application some internal settings can be tweaked like PARALLELIZE or VERBOSE.</sub>

# What does my solution do successfully
- Reads the .csv file into memory
- Distinguishes between static and dynamic cells
- Computes the dependencies for each cell (for the _^^_, _(A..Z)n_ and _E^_ operators)
- Splits the cells into independent buckets (this was required to be able to achieve concurrency)
- Launches a _goroutine_ to process each of the buckets (this part was actually pretty easy).
- Processes some of the operators (_^^_, _(A..Z)n_ and _E^_)
- Prints the results computed (or partially computed) to an output file

# What does my solution NOT do
- It does not correctly compute the dependencies for some operators (_A^v_, _@label<n>_ and function calls)
- It does not correctly process all the operators shown in the expressions (_A^v_, _@label<n>_ and function calls)
- It has no notion of labels (they're just static cells like the others)
- It does not detected circular dependencies

# Other remarks
- About 30h of effort were put into this project.
- Some parts of the code could probably be more efficient.
- Some parts of the code could probably use some more testing.
- All the code is documented - because that's how I like to code.
