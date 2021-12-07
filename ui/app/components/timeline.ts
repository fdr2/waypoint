import Component from '@glimmer/component';
import { inject as service } from '@ember/service';
import RouterService from '@ember/routing/router-service';
import { DeploymentExtended, BuildExtended, ReleaseExtended } from 'waypoint/services/api';
import ApiService from 'waypoint/services/api';
import { Status } from 'waypoint-pb';

interface Artifact {
  sequence: number;
  type: string;
  route: string;
  status?: Status.AsObject;
  isCurrentRoute: boolean;
}

export interface TimelineModel {
  build?: TimelineArtifact;
  deployment?: TimelineArtifact;
  release?: TimelineArtifact;
}
interface TimelineArtifact {
  sequence: number;
  status: Status.AsObject | undefined;
}
interface TimelineArgs {
  currentArtifact: DeploymentExtended | BuildExtended | ReleaseExtended;
  timeline: TimelineModel;
}

let TYPE_TRANSLATIONS = {
  build: 'page.artifact.timeline.build',
  deployment: 'page.artifact.timeline.deployment',
  release: 'page.artifact.timeline.release',
};

let TYPE_ROUTES = {
  build: 'workspace.projects.project.app.build',
  deployment: 'workspace.projects.project.app.deployment.deployment-seq',
  release: 'workspace.projects.project.app.release',
};

export default class Timeline extends Component<TimelineArgs> {
  @service api!: ApiService;
  @service router!: RouterService;

  areWeHere(currentArtifactKey: string): boolean {
    let entry = Object.entries(TYPE_ROUTES).find(([_, value]) =>
      this.router.currentRouteName.includes(value)
    );

    if (entry) {
      return entry[0] === currentArtifactKey;
    }
    return false;
  }

  get artifacts(): Artifact[] {
    let artifactsList: Artifact[] = [];
    for (let key in this.args.timeline) {
      let artifactObj = {
        sequence: this.args.timeline[key].sequence,
        type: TYPE_TRANSLATIONS[key],
        route: TYPE_ROUTES[key],
        status: this.args.timeline[key].status,
        isCurrentRoute: this.areWeHere(key),
      } as Artifact;
      artifactsList.push(artifactObj);
    }
    return artifactsList;
  }
}
