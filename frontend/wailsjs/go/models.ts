export namespace app {
	
	export class SessionResult {
	    sessionId: number;
	    textId: number;
	    summary: stats.Summary;
	    mistakes: typing.WordError[];
	    average: store.Averages;
	
	    static createFrom(source: any = {}) {
	        return new SessionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.textId = source["textId"];
	        this.summary = this.convertValues(source["summary"], stats.Summary);
	        this.mistakes = this.convertValues(source["mistakes"], typing.WordError);
	        this.average = this.convertValues(source["average"], store.Averages);
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
	export class TextInput {
	    title: string;
	    content: string;
	    language: string;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new TextInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.content = source["content"];
	        this.language = source["language"];
	        this.tags = source["tags"];
	    }
	}

}

export namespace stats {
	
	export class Summary {
	    durationMs: number;
	    keystrokes: number;
	    correctKeystrokes: number;
	    backspaces: number;
	    correctChars: number;
	    uncorrectedErrors: number;
	    wpmGross: number;
	    wpmNet: number;
	    cpm: number;
	    accuracy: number;
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.durationMs = source["durationMs"];
	        this.keystrokes = source["keystrokes"];
	        this.correctKeystrokes = source["correctKeystrokes"];
	        this.backspaces = source["backspaces"];
	        this.correctChars = source["correctChars"];
	        this.uncorrectedErrors = source["uncorrectedErrors"];
	        this.wpmGross = source["wpmGross"];
	        this.wpmNet = source["wpmNet"];
	        this.cpm = source["cpm"];
	        this.accuracy = source["accuracy"];
	    }
	}

}

export namespace store {
	
	export class Averages {
	    sessions: number;
	    wpmNet: number;
	    accuracy: number;
	
	    static createFrom(source: any = {}) {
	        return new Averages(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessions = source["sessions"];
	        this.wpmNet = source["wpmNet"];
	        this.accuracy = source["accuracy"];
	    }
	}
	export class Text {
	    id: number;
	    title: string;
	    content: string;
	    language: string;
	    source: string;
	    tags: string[];
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Text(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.language = source["language"];
	        this.source = source["source"];
	        this.tags = source["tags"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class TextSummary {
	    id: number;
	    title: string;
	    language: string;
	    source: string;
	    tags: string[];
	    length: number;
	    preview: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new TextSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.language = source["language"];
	        this.source = source["source"];
	        this.tags = source["tags"];
	        this.length = source["length"];
	        this.preview = source["preview"];
	        this.createdAt = source["createdAt"];
	    }
	}

}

export namespace typing {
	
	export class Keystroke {
	    seq: number;
	    position: number;
	    typed: string;
	    timeMs: number;
	
	    static createFrom(source: any = {}) {
	        return new Keystroke(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seq = source["seq"];
	        this.position = source["position"];
	        this.typed = source["typed"];
	        this.timeMs = source["timeMs"];
	    }
	}
	export class WordError {
	    expected: string;
	    typed: string;
	    position: number;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new WordError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.expected = source["expected"];
	        this.typed = source["typed"];
	        this.position = source["position"];
	        this.type = source["type"];
	    }
	}

}

