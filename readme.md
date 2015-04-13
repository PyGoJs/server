Server for a small school project that aims to provide a way for students to check into classes with their school card, without teachers having to set them as 'attended' manually.

This project will not be used in production. It's just for learning purposes for the creators. 

##Tasks the server is responsible for
###Check-in's
* Providing clients with information about the current (or next) class in a classroom. See HTTP handler /nextclass .
* Set a student as attended for the current class, when a client requests it. See HTTP handler /checkin .
    * Server has a list of clients containing the facility (classroom) the client is located at. 
    * Server needs to know the current schedule for classes.
    * Server needs to make sure the schedules are up to date (as far as a 30 minute cache can be called up to date).
    * Server has a list of students (and creates attendees) and classes (and creates lessons/class_items).

###API
* Give the website access to information about classes, lessons and attendee's.
    * See HTTP handler /api/class, /api/class_item and /api/attendee .
* Lock the attendee information for the website behind basic authentication (login).

##Errors given from /checkin for the client
1. Student not found
2. Too long till next clas
3. Student is already attending his/her current (or next < 15 minutes) class.
4. No class for today anymore (schedule item not found) &#42;
5. Student tries to check into the wrong class (example: student's class is in BA6.43, but he/she tries to check into BA6.49)

&#42; Also given from /nextclass
