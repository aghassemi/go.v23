/* Package ipc defines the public interface for all interprocess communication.

There are two actors in the system, clients and servers.  Clients invoke methods
on Servers, using the StartCall method provided by the Client interface.
Servers implement methods on named objects.  The named object is found using a
Dispatcher, and the method is invoked using an Invoker.

Instances of the Runtime host Clients and Servers, such instances may
simultaneously host both Clients and Servers.  The Runtime allows multiple names
to be simultaneously supported via the Dispatcher interface.

The naming package provides a rendezvous mechanism for Clients and Servers.  In
particular, it allows Runtimes hosting Servers to share Endpoints with Clients
that enables communication between them.  Endpoints encode sufficient addressing
information to enable communication.
*/
package ipc
