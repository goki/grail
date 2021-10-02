# grail

Grail is a (planned) email app in the GoGi GUI framework.

Just bookmarking the name for now.  Grail rhymes with mail, and has the same GoGi family sound.. and your golden challice of messages is overflowing or something like that.

Once the glide html rendering engine is working, then a natural next step would be an email app.  Presumably the relevant protocols are supported in some Go library, maybe even the standard library?  So then a basic email app seems relatively easy, and could provide a highly keyboard-centic efficient interface for email processing.  For example, the mac Mail.app does not natively support filing emails into folders through keyboard actions -- does everyone seriously do this using the mouse?  I use the MsgFiler add-on, together with the Keyboard Maestro keymapping app to make this happen, but it is still a pain every time they update the app, and searching has been rather flaky too.

Still, as usual, it is not like it is actually *worth it* at a purely rational level to write a whole email app, but once all the pieces are there, it is just too compelling to resist...

With grail and glide in place, together with gide and grid, all of the core desktop productivity apps would be available within this one framework, at least for latex / markdown based document management.  You might need a spreadsheet at some point, but maybe a line can be drawn there..

# Design

Keep-it-simple design, basically just like mac Mail.app, with a TreeView folder view on the left, message list next, and then message display, with popup window for composing new messages.

Key features:

* Remove attachments that actually works

* keyboard-based filing of messages

* Maybe some kind of markdown-based simple html formatter -- you write markdown, it translates to html and sends?  That would be nice.

* thread on markdown format: https://softwarerecs.stackexchange.com/questions/1022/email-client-that-supports-markdown  MailMate is most directly supported case and seems to have a similar philosophy as grail..  can have option to not even translate md -> HTML and just send the raw md -- this client and perhaps others will do the right thing..
* 
* 
