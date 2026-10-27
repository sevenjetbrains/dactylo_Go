export namespace app {
	
	export class SystemInfo {
	    version: string;
	    platform: string;
	    databasePath: string;
	    schemaVersion: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.platform = source["platform"];
	        this.databasePath = source["databasePath"];
	        this.schemaVersion = source["schemaVersion"];
	    }
	}

}

