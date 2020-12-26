class SystemdModel{
    constructor(hostaddlistener) {
        // hostname => hostmodel
        this.map = new Map()
        this.hostaddlistener = hostaddlistener
        this.es = new EventSource("events?stream=messages");
        this.es.onmessage = (msg) => {
            const unit = JSON.parse(msg.data);
            this.update(unit);
        }
    }
    update(status) {
        var hostmodel
        const key = status.host
        if (this.map.has(key)) {
            hostmodel = this.map.get(key)
        } else {
            hostmodel = new HostModel(key)
            this.map.set(key, hostmodel)
            if (this.hostaddlistener) {
                this.hostaddlistener(key)
            }
        }
        hostmodel.update(status)
    }
    getHost(hostname) {
        return this.map.get(hostname)
    }
    hosts() {
        return this.map.values()
    }
}

class HostModel {
    constructor(hostname) {
        this.hostname = hostname
        this.statuslistener = null
        // unitname => unitmodel
        this.map = new Map()
    }
    update(status) {
        const key = status.Name
        const added = !this.map.has(key)
        this.map.set(key, status)
        if (this.statuslistener) {
            this.statuslistener(status, added)
        }
    }
    units() {
        return this.map.values()
    }
}