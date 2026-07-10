export namespace config {
	
	export class AutoPickSettings {
	    enabled: boolean;
	    primaryChamp: number;
	    secondaryChamp: number;
	    champPool: number[];
	
	    static createFrom(source: any = {}) {
	        return new AutoPickSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.primaryChamp = source["primaryChamp"];
	        this.secondaryChamp = source["secondaryChamp"];
	        this.champPool = source["champPool"];
	    }
	}
	export class Settings {
	    schemaVersion: number;
	    autoAccept: boolean;
	    autoPick: AutoPickSettings;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.autoAccept = source["autoAccept"];
	        this.autoPick = this.convertValues(source["autoPick"], AutoPickSettings);
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

export namespace lcu {
	
	export class SummonerInfo {
	    displayName: string;
	    summonerLevel: number;
	    profileIconId: number;
	
	    static createFrom(source: any = {}) {
	        return new SummonerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.displayName = source["displayName"];
	        this.summonerLevel = source["summonerLevel"];
	        this.profileIconId = source["profileIconId"];
	    }
	}

}

export namespace main {
	
	export class LCUResponse {
	    isActive: boolean;
	    port: string;
	    locale: string;
	    gameVersion: string;
	    summoner?: lcu.SummonerInfo;
	
	    static createFrom(source: any = {}) {
	        return new LCUResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isActive = source["isActive"];
	        this.port = source["port"];
	        this.locale = source["locale"];
	        this.gameVersion = source["gameVersion"];
	        this.summoner = this.convertValues(source["summoner"], lcu.SummonerInfo);
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

