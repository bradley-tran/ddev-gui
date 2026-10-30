export namespace backend {
	
	export class AppVersionInfo {
	    version: string;
	    commitHash: string;
	
	    static createFrom(source: any = {}) {
	        return new AppVersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.commitHash = source["commitHash"];
	    }
	}
	export class WindowBounds {
	    width: number;
	    height: number;
	    x: number;
	    y: number;
	    maximized: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WindowBounds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.maximized = source["maximized"];
	    }
	}

}

