export namespace common {
	
	export class ProxyAuth {
	    Username: string;
	    Password: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyAuth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Username = source["Username"];
	        this.Password = source["Password"];
	    }
	}

}

export namespace main {
	
	export class ListenerStat {
	    Sent: number;
	    Received: number;
	
	    static createFrom(source: any = {}) {
	        return new ListenerStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Sent = source["Sent"];
	        this.Received = source["Received"];
	    }
	}
	export class ServerFilter {
	    Tags: string[];
	    IgnoreAll: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ServerFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Tags = source["Tags"];
	        this.IgnoreAll = source["IgnoreAll"];
	    }
	}
	export class LocalListener {
	    Port: number;
	    Listener: any;
	    Auth?: common.ProxyAuth;
	    Filter: ServerFilter;
	    Stat: ListenerStat;
	
	    static createFrom(source: any = {}) {
	        return new LocalListener(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Port = source["Port"];
	        this.Listener = source["Listener"];
	        this.Auth = this.convertValues(source["Auth"], common.ProxyAuth);
	        this.Filter = this.convertValues(source["Filter"], ServerFilter);
	        this.Stat = this.convertValues(source["Stat"], ListenerStat);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ManagedLocalListener {
	    Listener?: LocalListener;
	    IsServing: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ManagedLocalListener(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Listener = this.convertValues(source["Listener"], LocalListener);
	        this.IsServing = source["IsServing"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ManagedProxyServer {
	    Server?: proxyserver.Server;
	    Tags: Record<string, boolean>;
	
	    static createFrom(source: any = {}) {
	        return new ManagedProxyServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Server = this.convertValues(source["Server"], proxyserver.Server);
	        this.Tags = source["Tags"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class listenerServerManager {
	    Listeners: Record<number, ManagedLocalListener>;
	    Servers: Record<string, ManagedProxyServer>;
	    IsServing: boolean;
	    // Go type: sync
	    Wg: any;
	
	    static createFrom(source: any = {}) {
	        return new listenerServerManager(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Listeners = this.convertValues(source["Listeners"], ManagedLocalListener, true);
	        this.Servers = this.convertValues(source["Servers"], ManagedProxyServer, true);
	        this.IsServing = source["IsServing"];
	        this.Wg = this.convertValues(source["Wg"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace proxyserver {
	
	export class Server {
	    Host: string;
	    Port: number;
	    Auth?: common.ProxyAuth;
	    Timeout: number;
	    PublicIp: string;
	    Latency: number;
	    // Go type: time
	    LastChecked?: any;
	    Protocols: Record<string, boolean>;
	
	    static createFrom(source: any = {}) {
	        return new Server(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.Auth = this.convertValues(source["Auth"], common.ProxyAuth);
	        this.Timeout = source["Timeout"];
	        this.PublicIp = source["PublicIp"];
	        this.Latency = source["Latency"];
	        this.LastChecked = this.convertValues(source["LastChecked"], null);
	        this.Protocols = source["Protocols"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

