# FFWebGoAPI

This is the companion API for the FFPSWebApp project. The connection string to the database is set in variables.go, and the cors policy 
should be set there as well. 

This API pulls a lot of data from json files that were obtained directly from the game client. It also used the lua scripts that are part 
of the private FFXI server code. Neither of these have been included in this repo. 

Parsing through these scripts and using details from the json files made it possible to retrieve very specific data about different aspects of the game.
These lua scripts are available with any version of the server code. In this API the scripts folder was placed in the root directory where 
the main go file is.

I made this for fun and because I loved playing FFXI in it's prime. This was my first time writing anything in Go, so the code may not be pretty, but it does work when I tested it in Feb 2023 with newer Wings database version. 

If anyone would like more information about the structure of the folders or would like the images and json files that are not included with the 
overall WebApp, you can contact me at gteam6@protonmail.com.


