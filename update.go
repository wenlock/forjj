package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"regexp"
)

// Execute an update on the workspace given.
//
// Workspace data has been initialized or loaded.
// forjj-options has been initialized or loaded
func (a *Forj) Update() error {
	if _, err := a.w.Ensure_exist(); err != nil {
		return fmt.Errorf("Invalid workspace. %s. Please create it with 'forjj create'", err)
	}

	defer func() {
		// save infra repository location in the workspace.
		a.w.Save()

		if err := a.s.Save(); err != nil {
			log.Printf("%s", err)
		}
	}()

	if err := a.define_infra_upstream(); err != nil {
		return fmt.Errorf("Unable to identify a valid infra repository upstream. %s", err)
	}

	gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

	a.DefineDefaultUpstream()

	// missing:true to check if some required values are missing.
	if err := a.ScanAndSetObjectData(true) ; err != nil {
		return fmt.Errorf("Unable to update. %s", err)
	}

	// Checking infra repository: A valid infra repo is a git repository with at least one commit and
	// a Forjfile from repo root.
	if err := a.i.Use(a.f.InfraPath()) ; err != nil {
		return fmt.Errorf("Failed to update your infra repository. %s", err)
	}

	// Now, we are in the infra repo root directory and at least, the 1st commit exist.

	// TODO: flow_start to execute instructions before updating source code for existing apps in appropriate branch. Possible if a flow is already implemented otherwise git must stay in master branch
	// flow_start()

	// Disabled as not ready.
	//if err := a.MoveToFixBranch(*a.Actions["update"].argsv["branch"]) ; err != nil {
	//    return fmt.Errorf("Unable to move to your feature branch. %s", err)
	//}

	instances := a.define_drivers_execution_order()

	// Loop on drivers requested like github or jenkins
	for _, instance := range instances {
		d := a.drivers[instance]
		if err, aborted := a.do_driver_task("update", instance); err != nil {
			if !aborted {
				return fmt.Errorf("Failed to update '%s' source files. %s", instance, err)
			}
			log.Printf("Warning. %s", err)
			continue
		}

		if d.HasNoFiles() {
			gotrace.Info("No files to add/commit.")
			continue
		}

		// Committing source code.
		if err := a.do_driver_add(d); err != nil {
			return fmt.Errorf("Failed to Add '%s' source files. %s", instance, err)
		}
	}
	/*	// If the upstream driver has updated his source, we need to get and commit them. If
		// Commiting source code.
		if d, found := a.drivers[a.w.Instance]; no_new_infra && found {
			if err := a.do_driver_commit(d); err != nil {
				return fmt.Errorf("Failed to commit '%s' source files. %s", a.w.Instance, err)
			}
		}

		// a.o.update_options()

		// Save&add forjj-repos, save&add forjj-options & then commit
		defer func() {
			// Save forjj-repos.yml
			if err := a.RepoCodeSave(); err != nil {
				log.Printf("%s", err)
			}
			if err := a.SaveForjjPluginsOptions(); err != nil {
				log.Printf("%s", err)
			}

			// Save forjj-options.yml
			a.SaveForge(fmt.Sprintf("Organization %s updated.", a.w.Organization))
			log.Printf("As soon as you are happy with your fixes, do a git push to submit your collection of fixes related to '%s' to your team.", a.Branch)
		}()

		// Loop on drivers requested like jenkins classified as ci type.
		for instance, d := range a.drivers {

			if instance == a.w.Instance {
				continue // Do not try to update infra-upstream twice.
			}

			repos_num := a.GetReposRequestedFor(instance, "update")
			gotrace.Trace("Instance '%s' hosts %s.", instance, NumReposDisplay(repos_num))
			if !d.AppRequest() && repos_num == 0 {
				continue // Do not try to update a non requested app (--apps) or an instance having no requested repo updates.
			}

			if err, aborted := a.do_driver_task("update", instance); err != nil {
				if !aborted {
					return fmt.Errorf("Failed to update '%s' source files. %s", instance, err)
				}
				log.Printf("Warning. %s", err)
			}

			// Committing source code.
			if err := a.do_driver_commit(d); err != nil {
				return fmt.Errorf("Failed to commit '%s' source files. %s", instance, err)
			}

		}

		// TODO: Implement flow_close() to close the create task
		// flow_close()*/
	return nil
}

func (a *Forj) MoveToFixBranch(branch string) error {
	a.Branch = branch

	if ok, _ := regexp.MatchString(`^[\w_-]+$`, branch); !ok {
		return fmt.Errorf("Invalid git branch name '%s'. alphanumeric, '_' and '-' are supported.", branch)
	}
	return nil
}
