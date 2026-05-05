export namespace main {
	
	export class BuildParams {
	    mode: string;
	    langs: string[];
	    langTyps: {[key: string]: string};
	    rootTyp: string;
	    cols: string;
	    media: string;
	    audience: string;
	    production: boolean;
	    coverImage: string;
	    extraArgs: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.langs = source["langs"];
	        this.langTyps = source["langTyps"];
	        this.rootTyp = source["rootTyp"];
	        this.cols = source["cols"];
	        this.media = source["media"];
	        this.audience = source["audience"];
	        this.production = source["production"];
	        this.coverImage = source["coverImage"];
	        this.extraArgs = source["extraArgs"];
	    }
	}
	export class BuildResult {
	    success: boolean;
	    log: string;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new BuildResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.log = source["log"];
	        this.errors = source["errors"];
	    }
	}
	export class ScanResult {
	    rootTypes: string[];
	    langFolders: string[];
	    langTypes: {[key: string]: string[]};
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rootTypes = source["rootTypes"];
	        this.langFolders = source["langFolders"];
	        this.langTypes = source["langTypes"];
	    }
	}

}

