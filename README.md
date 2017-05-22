# Raven chat
For birds by birds. This chat is a concept by moi, your silly corvid pal. The idea is growing chat networks based on
client locale. E.g. assume you are the only one connected to the chat program. initially the chat you are in is
restricted to a region that contains only you. Then, another user connects one region over. The chat application will
join your regions together and now the network is larger. However, it can shrink as well. Should a network of three
regions joined end-to-end be split by removal of the region in the middle, the two remaining regions will be isolated.

This fun little concept is something I had been sitting on for a while and I think it will be a fun concept to code
since it involves a bit of graph theory (mostly basic stuff) and the concept should prove fun as a general means of
communication. Furthermore it improves my lock-free programming skills which is also a plus yo.

# client addition
We add clients in a pipeline format. The addition of a client is as follows:

* Add client to a region
* Check if the client added means a new region added to a network
* If client added a new region, check to see if it connects networks
