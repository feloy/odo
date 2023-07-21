import { Component, OnInit } from '@angular/core';
import { DevstateService } from './services/devstate.service';
import { DomSanitizer } from '@angular/platform-browser';
import { MermaidService } from './services/mermaid.service';
import { StateService } from './services/state.service';
import { MatIconRegistry } from "@angular/material/icon";
import { OdoapiService } from './services/odoapi.service';
import { SseService } from './services/sse.service';
import {DevfileContent} from "./api-gen";
import { TelemetryResponse } from './api-gen';
import { MatTabChangeEvent } from '@angular/material/tabs';
import { TelemetryService } from './services/telemetry.service';
import { MatSnackBar, MatSnackBarRef } from '@angular/material/snack-bar';
import { ConfirmComponent } from './components/confirm/confirm.component';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {

  protected tabNames: string[] = [
    "YAML",
    "Chart",
    "Metadata",
    "Commands",
    "Events",
    "Containers",
    "Images",
    "Resources",
    "Volumes"
  ];
  protected mermaidContent: string = "";
  protected devfileYaml: string = "";
  protected errorMessage: string  = "";
  private snackBarRef: MatSnackBarRef<ConfirmComponent> | null = null;

  constructor(
    protected sanitizer: DomSanitizer,
    private matIconRegistry: MatIconRegistry,
    private wasmGo: DevstateService,
    private odoApi: OdoapiService,
    private mermaid: MermaidService,
    private state: StateService,
    private sse: SseService,
    private telemetry: TelemetryService,
    private snackbar: MatSnackBar
  ) {
    this.matIconRegistry.addSvgIcon(
      `github`,
      this.sanitizer.bypassSecurityTrustResourceUrl(`../assets/github-24.svg`)
    );
  }

  ngOnInit() {
    const loading = document.getElementById("loading");
    if (loading != null) {
      loading.style.visibility = "hidden";
    }

    const devfile = this.odoApi.getDevfile();
    devfile.subscribe({
      next: (devfile) => {
        if (devfile.content != undefined) {
          this.propagateChange(devfile.content, false);
        }
      }
    });

    this.state.state.subscribe(async newContent => {
      if (newContent == null) {
        return;
      }

      this.devfileYaml = newContent.content;

      const result = this.wasmGo.getFlowChart();
      result.subscribe({
        next: async (res) => {
          const svg = await this.mermaid.getMermaidAsSVG(res.chart);
          this.mermaidContent = svg;      
        },
        error: (error) => {
          console.log(error);
        }
      });
    });

    this.sse.subscribeTo(['DevfileUpdated']).subscribe(event => {
      if (this.snackBarRef != null) {
        this.snackBarRef.afterDismissed().subscribe(() => {});
        this.snackBarRef.dismiss();
      }
      this.snackBarRef = this.snackbar.openFromComponent(ConfirmComponent, { data: { 
        message: "The Devfile has changed on disk. Do you want to update it here?",
        noLabel: "Cancel", 
        yesLabel: "Update"
      }});
      this.snackBarRef.onAction().subscribe(() => {
        let newDevfile: DevfileContent = JSON.parse(event.data);
        if (newDevfile.content != undefined) {
          this.propagateChange(newDevfile.content, false);
        }
        this.snackBarRef = null;
      });
      this.snackBarRef.afterDismissed().subscribe(() => { 
        this.snackBarRef = null;
      });
    });

    this.odoApi.telemetry().subscribe({
      next: (data: TelemetryResponse) => {
        if (data.enabled) {
          if (data.apikey == null || data.userid == null) {
            return;
          }
          this.telemetry.init(data.apikey, data.userid)
          this.telemetry.track("[ui] start");
        }    
      },
      error: () => {}
    })
  }

  propagateChange(content: string, saveToApi: boolean){
    const result = this.wasmGo.setDevfileContent(content);
    result.subscribe({
      next: (value) => {
        this.errorMessage = '';
        this.state.changeDevfileYaml(value);
        if (saveToApi) {
          this.odoApi.saveDevfile(value.content).subscribe({
            next: () => {},
            error: (error) => {
              this.errorMessage = error.error.message;
            }
          });
        }
      },
      error: (error) => {
        this.errorMessage = error.error.message;
      }
    });
  }

  onSave(content: string) {
    this.telemetry.track("[ui] save devfile to disk");
    this.propagateChange(content, true);
  }

  onApply(content: string) {
    this.telemetry.track("[ui] change devfile from textarea");
    this.propagateChange(content, false);
  }

  clear() {
    if (confirm('You will delete the content of the Devfile. Continue?')) {
      this.telemetry.track("[ui] clear devfile");
      this.wasmGo.clearDevfileContent().subscribe({
        next: (value) => {
          this.propagateChange(value.content, false);
        }
      });
    }
  }

  onSelectedTabChange(e: MatTabChangeEvent) {
    this.telemetry.track("[ui] change to tab "+this.tabNames[e.index]);
  }
}
