import time
import threading
try:
    from greenlet import getcurrent as get_ident
except ImportError:
    try:
        from thread import get_ident
    except ImportError:
        from _thread import get_ident

class RpcEvent(object):
    """
    An Event-like class that signals all active clients when a new RPC object
    is available.
    """
    def __init__(self):
        self.events = {}

    def wait(self):
        """Invoked from each client's thread to wait for the next RPC object."""
        ident = get_ident()
        if ident not in self.events:
            self.events[ident] = [threading.Event(), time.time()]
        return self.events[ident][0].wait()

    def set(self):
        """Invoked by the RTD thread when a new RPC object is available."""
        now = time.time()
        remove = None
        for ident, event in self.events.items():
            if not event[0].isSet():
                event[0].set()
                event[1] = now
            else:
                if now - event[1] > 5:
                    remove = ident
        if remove:
            del self.events[remove]

    def clear(self):
        """Invoked from each client's thread after an RPC object was processed."""
        self.events[get_ident()][0].clear()


class BaseRPC(object):
    thread = None
    rtd = None
    last_access = 0
    event = RpcEvent()

    def __init__(self):
        if BaseRPC.thread is None:
            BaseRPC.last_access = time.time()
            BaseRPC.thread = threading.Thread(target=self._thread)
            BaseRPC.thread.start()

            while self.get_rtd() is None:
                time.sleep(0)

    def get_rtd(self):
        BaseRPC.last_access = time.time()
        BaseRPC.event.wait()
        BaseRPC.event.clear()
        return BaseRPC.rtd

    @staticmethod
    def frames():
        raise RuntimeError('Must be implemented by subclasses')

    @classmethod
    def _thread(cls):
        """RPC background thread"""
        print('Starting RPC thread')
        rtd_iterator = cls.frames()
        for rtd in rtd_iterator:
            BaseRPC.rtd = rtd
            BaseRPC.event.set()
            time.sleep(0)

            if time.time() - BaseRPC.last_access > 20:
                rtd_iterator.close()
                print('Stopping FMT thread due to inactivity')
                break
        BaseRPC.thread = None
