import { Component, Input } from '@angular/core';
import { MatCheckboxChange } from '@angular/material/checkbox';
import { StateService } from 'src/app/services/state.service';
import { Command, DevstateService } from 'src/app/services/devstate.service';

@Component({
  selector: 'app-commands-list',
  templateUrl: './commands-list.component.html',
  styleUrls: ['./commands-list.component.css']
})
export class CommandsListComponent {
  @Input() commands: Command[] | undefined;
  @Input() kind: string = "";
  @Input() dragDisabled: boolean = true;

  constructor(
    private devstate: DevstateService,
    private state: StateService,
  ) {}

  toggleDefault(event: MatCheckboxChange, command: string, group: string) {
    if (event.checked) {
      this.setDefault(command, group);
    } else {
      this.unsetDefault(command);
    }
  }

  setDefault(command: string, group: string) {
    const result = this.devstate.setDefaultCommand(command, group);
    result.subscribe({
      next: (value) => {
        this.state.changeDevfileYaml(value);
      }, 
      error: (error) => {
        alert(error.error.message);
      }
    });
  }

  unsetDefault(command: string) {
    const result = this.devstate.unsetDefaultCommand(command);
    result.subscribe({
      next: (value) => {
        this.state.changeDevfileYaml(value);
      }, 
      error: (error) => {
        alert(error.error.message);
      }
    });
  }

  getCommandsByKind(commands: Command[] | undefined, kind: string ): Command[] | undefined {
    return commands?.filter((c: Command) => c.group == kind);
  }

  delete(command: string) {
    if(confirm('You will delete the command "'+command+'". Continue?')) {
      const result = this.devstate.deleteCommand(command);
      result.subscribe({
        next: (value) => {
          this.state.changeDevfileYaml(value);
        },
        error: (error) => {
          alert(error.error.message);
        }
      });
    }
  }
}
