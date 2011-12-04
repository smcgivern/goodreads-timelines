This is a super-hacky (for instance, there are no tests!) mini-site
that pulls a user's books from [Goodreads][goodreads] and displays a
chart of reading per month / day, along with a calendar of when they
finished reading each book.

Goodreads store a lot of fields about books but all this really cares
about is the date that the person read the book on, because that's the
only field on my profile that's semi-reliable. I don't know what
happens if a book's been read more than once -- either in the API or
in this tool.

[goodreads]: http://www.goodreads.com/
