import pytest
from .common import random_str
import time
from rancher import ApiError


def ns_default_quota():
    return {"limit": {"pods": "4"}}


def ns_quota():
    return {"limit": {"pods": "4"}}


def ns_small_quota():
    return {"limit": {"pods": "1"}}


def ns_large_quota():
    return {"limit": {"pods": "200"}}


def default_project_quota():
    return {"limit": {"pods": "100"}}


def ns_default_limit():
    return {"requestsCpu": "1",
            "requestsMemory": "1Gi",
            "limitsCpu": "2",
            "limitsMemory": "2Gi"}


def wait_for_applied_quota_set(admin_cc_client, ns, timeout=30):
    start = time.time()
    ns = admin_cc_client.reload(ns)
    a = ns.annotations["cattle.io/status"]
    while a is None:
        time.sleep(.5)
        ns = admin_cc_client.reload(ns)
        a = ns.annotations["cattle.io/status"]
        if time.time() - start > timeout:
            raise Exception('Timeout waiting for'
                            ' resource quota to be validated')

    while "Validating resource quota" in a or "exceeds project limit" in a:
        time.sleep(.5)
        ns = admin_cc_client.reload(ns)
        a = ns.annotations["cattle.io/status"]
        if time.time() - start > timeout:
            raise Exception('Timeout waiting for'
                            ' resource quota to be validated')


def test_namespace_resource_quota(admin_cc, admin_pc):
    q = default_project_quota()
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=q,
                              namespaceDefaultResourceQuota=q)

    p = admin_cc.management.client.wait_success(p)
    assert p.resourceQuota is not None

    client = admin_pc.cluster.client
    ns = client.create_namespace(name=random_str(),
                                 projectId=p.id,
                                 resourceQuota=ns_quota())
    assert ns is not None
    assert ns.resourceQuota is not None
    wait_for_applied_quota_set(admin_pc.cluster.client,
                               ns)


def test_project_resource_quota_fields(admin_cc):
    q = default_project_quota()
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=q,
                              namespaceDefaultResourceQuota=q)
    p = admin_cc.management.client.wait_success(p)

    assert p.resourceQuota is not None
    assert p.resourceQuota.limit.pods == '100'
    assert p.namespaceDefaultResourceQuota is not None
    assert p.namespaceDefaultResourceQuota.limit.pods == '100'


def test_resource_quota_ns_create(admin_cc, admin_pc):
    q = default_project_quota()
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=q,
                              namespaceDefaultResourceQuota=q)
    p = admin_cc.management.client.wait_success(p)
    assert p.resourceQuota is not None
    assert p.resourceQuota.limit.pods == '100'

    ns = admin_pc.cluster.client.create_namespace(name=random_str(),
                                                  projectId=p.id,
                                                  resourceQuota=ns_quota())
    assert ns is not None
    assert ns.resourceQuota is not None
    wait_for_applied_quota_set(admin_pc.cluster.client, ns)


def test_default_resource_quota_ns_set(admin_cc, admin_pc):
    q = ns_default_quota()
    pq = default_project_quota()
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=pq,
                              namespaceDefaultResourceQuota=q)
    assert p.resourceQuota is not None
    assert p.namespaceDefaultResourceQuota is not None

    ns = admin_pc.cluster.client.create_namespace(name=random_str(),
                                                  projectId=p.id)
    wait_for_applied_quota_set(admin_pc.cluster.client,
                               ns)


def test_quota_ns_create_exceed(admin_cc, admin_pc):
    q = default_project_quota()
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=q,
                              namespaceDefaultResourceQuota=q)
    p = admin_cc.management.client.wait_success(p)
    assert p.resourceQuota is not None

    # namespace quota exceeding project resource quota
    cluster_client = admin_pc.cluster.client
    with pytest.raises(ApiError) as e:
        cluster_client.create_namespace(name=random_str(),
                                        projectId=p.id,
                                        resourceQuota=ns_large_quota())

    assert e.value.error.status == 422


def test_default_resource_quota_ns_create_invalid_combined(admin_cc, admin_pc):
    pq = default_project_quota()
    q = ns_default_quota()
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=pq,
                              namespaceDefaultResourceQuota=q)
    p = admin_cc.management.client.wait_success(p)
    assert p.resourceQuota is not None
    assert p.namespaceDefaultResourceQuota is not None

    ns = admin_pc.cluster.client.create_namespace(name=random_str(),
                                                  projectId=p.id,
                                                  resourceQuota=ns_quota())
    assert ns is not None
    assert ns.resourceQuota is not None

    # namespace quota exceeding project resource quota
    cluster_client = admin_pc.cluster.client
    with pytest.raises(ApiError) as e:
        cluster_client.create_namespace(name=random_str(),
                                        projectId=p.id,
                                        resourceQuota=ns_large_quota())

    assert e.value.error.status == 422

    # quota within limits
    ns = cluster_client.create_namespace(name=random_str(),
                                         projectId=p.id,
                                         resourceQuota=ns_small_quota())
    assert ns is not None
    assert ns.resourceQuota is not None
    wait_for_applied_quota_set(admin_pc.cluster.client, ns, 10)
    ns = admin_cc.client.reload(ns)

    # update namespace with exceeding quota
    with pytest.raises(ApiError) as e:
        admin_pc.cluster.client.update(ns,
                                       resourceQuota=ns_large_quota())

    assert e.value.error.status == 422


def test_project_used_quota(admin_cc, admin_pc):
    pq = default_project_quota()
    q = ns_default_quota()
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=pq,
                              namespaceDefaultResourceQuota=q)
    p = admin_cc.management.client.wait_success(p)
    assert p.resourceQuota is not None
    assert p.namespaceDefaultResourceQuota is not None

    ns = admin_pc.cluster.client.create_namespace(name=random_str(),
                                                  projectId=p.id)
    wait_for_applied_quota_set(admin_pc.cluster.client,
                               ns)
    used = wait_for_used_pods_limit_set(admin_cc.management.client, p)
    assert used.pods == "4"


def wait_for_used_pods_limit_set(admin_cc_client, project, timeout=30,
                                 value="0"):
    start = time.time()
    project = admin_cc_client.reload(project)
    while "usedLimit" not in project.resourceQuota \
            or "pods" not in project.resourceQuota.usedLimit:
        time.sleep(.5)
        project = admin_cc_client.reload(project)
        if time.time() - start > timeout:
            raise Exception('Timeout waiting for'
                            ' project.usedLimit.pods to be set')
    if value == "0":
        return project.resourceQuota.usedLimit
    while project.resourceQuota.usedLimit.pods != value:
        time.sleep(.5)
        project = admin_cc_client.reload(project)
        if time.time() - start > timeout:
            raise Exception('Timeout waiting for'
                            ' project.usedLimit.pods to be set to ' + value)


def test_default_resource_quota_project_update(admin_cc, admin_pc):
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id)

    ns = admin_pc.cluster.client.create_namespace(name=random_str(),
                                                  projectId=p.id)
    wait_for_applied_quota_set(admin_pc.cluster.client, ns, 10)

    pq = default_project_quota()
    q = ns_default_quota()
    p = admin_cc.management.client.update(p,
                                          resourceQuota=pq,
                                          namespaceDefaultResourceQuota=q)
    assert p.resourceQuota is not None
    assert p.namespaceDefaultResourceQuota is not None
    wait_for_applied_quota_set(admin_pc.cluster.client,
                               ns)


def test_api_validation_project(admin_cc):
    client = admin_cc.management.client
    q = default_project_quota()
    # default namespace quota missing
    with pytest.raises(ApiError) as e:
        client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=q)

    assert e.value.error.status == 422

    # default namespace quota as None
    with pytest.raises(ApiError) as e:
        client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=q,
                              namespaceDefaultResourceQuota=None)

    assert e.value.error.status == 422

    with pytest.raises(ApiError) as e:
        client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              namespaceDefaultResourceQuota=q)

    assert e.value.error.status == 422

    lq = ns_large_quota()
    with pytest.raises(ApiError) as e:
        client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=q,
                              namespaceDefaultResourceQuota=lq)

    assert e.value.error.status == 422

    pq = {"limit": {"pods": "100",
                    "services": "100"}}
    iq = {"limit": {"pods": "100"}}
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id)
    with pytest.raises(ApiError) as e:
        admin_cc.management.client.update(p,
                                          resourceQuota=pq,
                                          namespaceDefaultResourceQuota=iq)


def test_api_validation_namespace(admin_cc, admin_pc):
    pq = {"limit": {"pods": "100",
                    "services": "100"}}
    q = {"limit": {"pods": "10",
                   "services": "10"}}
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=pq,
                              namespaceDefaultResourceQuota=q)
    p = admin_cc.management.client.wait_success(p)
    assert p.resourceQuota is not None
    assert p.namespaceDefaultResourceQuota is not None

    nsq = {"limit": {"pods": "10"}}
    with pytest.raises(ApiError) as e:
        admin_pc.cluster.client.create_namespace(name=random_str(),
                                                 projectId=p.id,
                                                 resourceQuota=nsq)
    assert e.value.error.status == 422


def test_used_quota_exact_match(admin_cc, admin_pc):
    pq = {"limit": {"pods": "10"}}
    q = {"limit": {"pods": "2"}}
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=pq,
                              namespaceDefaultResourceQuota=q)
    p = admin_cc.management.client.wait_success(p)
    assert p.resourceQuota is not None
    assert p.namespaceDefaultResourceQuota is not None

    nsq = {"limit": {"pods": "2"}}
    admin_pc.cluster.client.create_namespace(name=random_str(),
                                             projectId=p.id,
                                             resourceQuota=nsq)

    nsq = {"limit": {"pods": "8"}}
    admin_pc.cluster.client.create_namespace(name=random_str(),
                                             projectId=p.id,
                                             resourceQuota=nsq)
    wait_for_used_pods_limit_set(admin_cc.management.client, p, 10, "10")

    # try reducing the project quota
    pq = {"limit": {"pods": "8"}}
    dq = {"limit": {"pods": "1"}}
    with pytest.raises(ApiError) as e:
        admin_cc.management.client.update(p,
                                          resourceQuota=pq,
                                          namespaceDefaultResourceQuota=dq)
    assert e.value.error.status == 422


def test_add_remove_fields(admin_cc, admin_pc):
    pq = {"limit": {"pods": "10"}}
    q = {"limit": {"pods": "2"}}
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=pq,
                              namespaceDefaultResourceQuota=q)
    p = admin_cc.management.client.wait_success(p)
    assert p.resourceQuota is not None
    assert p.namespaceDefaultResourceQuota is not None

    nsq = {"limit": {"pods": "2"}}
    admin_pc.cluster.client.create_namespace(name=random_str(),
                                             projectId=p.id,
                                             resourceQuota=nsq)

    wait_for_used_pods_limit_set(admin_cc.management.client, p,
                                 10, "2")

    admin_pc.cluster.client.create_namespace(name=random_str(),
                                             projectId=p.id,
                                             resourceQuota=nsq)

    wait_for_used_pods_limit_set(admin_cc.management.client, p,
                                 10, "4")

    # update project with services quota
    with pytest.raises(ApiError) as e:
        pq = {"limit": {"pods": "10", "services": "10"}}
        dq = {"limit": {"pods": "2", "services": "7"}}
        admin_cc.management.client.update(p,
                                          resourceQuota=pq,
                                          namespaceDefaultResourceQuota=dq)
    assert e.value.error.status == 422

    pq = {"limit": {"pods": "10", "services": "10"}}
    dq = {"limit": {"pods": "2", "services": "2"}}
    p = admin_cc.management.client.update(p,
                                          resourceQuota=pq,
                                          namespaceDefaultResourceQuota=dq)
    wait_for_used_svcs_limit_set(admin_cc.management.client, p,
                                 30, "4")

    # remove services quota
    pq = {"limit": {"pods": "10"}}
    dq = {"limit": {"pods": "2"}}
    p = admin_cc.management.client.update(p,
                                          resourceQuota=pq,
                                          namespaceDefaultResourceQuota=dq)
    wait_for_used_svcs_limit_set(admin_cc.management.client, p,
                                 30, "0")


def wait_for_used_svcs_limit_set(admin_cc_client, project, timeout=30,
                                 value="0"):
    start = time.time()
    project = admin_cc_client.reload(project)
    while "usedLimit" not in project.resourceQuota \
            or "services" not in project.resourceQuota.usedLimit:
        time.sleep(.5)
        project = admin_cc_client.reload(project)
        if time.time() - start > timeout:
            raise Exception('Timeout waiting for'
                            ' project.usedLimit.services to be set')
    if value == "0":
        return project.resourceQuota.usedLimit
    while project.resourceQuota.usedLimit.services != value:
        time.sleep(.5)
        project = admin_cc_client.reload(project)
        if time.time() - start > timeout:
            raise Exception('Timeout waiting for'
                            ' usedLimit.services to be set to ' + value)


def test_update_quota(admin_cc, admin_pc):
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id)
    p = admin_cc.management.client.wait_success(p)

    # create 4 namespaces
    for x in range(4):
        admin_pc.cluster.client.create_namespace(name=random_str(),
                                                 projectId=p.id)

    # update project with default namespace quota that
    # won't qualify for the limit
    with pytest.raises(ApiError) as e:
        pq = {"limit": {"pods": "5"}}
        dq = {"limit": {"pods": "2"}}
        admin_cc.management.client.update(p,
                                          resourceQuota=pq,
                                          namespaceDefaultResourceQuota=dq)
    assert e.value.error.status == 422


def test_container_resource_limit(admin_cc, admin_pc):
    q = default_project_quota()
    lmt = ns_default_limit()
    client = admin_cc.management.client
    p = client.create_project(name='test-' + random_str(),
                              clusterId=admin_cc.cluster.id,
                              resourceQuota=q,
                              namespaceDefaultResourceQuota=q,
                              containerDefaultResourceLimit=lmt)

    p = admin_cc.management.client.wait_success(p)
    assert p.resourceQuota is not None
    assert p.containerDefaultResourceLimit is not None

    client = admin_pc.cluster.client
    ns = client.create_namespace(name=random_str(),
                                 projectId=p.id,
                                 resourceQuota=ns_quota())
    assert ns is not None
    assert ns.resourceQuota is not None
    assert ns.containerDefaultResourceLimit is not None
    wait_for_applied_quota_set(admin_pc.cluster.client,
                               ns)

    # reset the container limit
    p = admin_cc.management.client.update(p,
                                          containerDefaultResourceLimit=None)

    assert not hasattr(p, 'containerDefaultResourceLimit')

    ns = admin_cc.management.client.update(ns,
                                           containerDefaultResourceLimit=None)
    assert len(ns.containerDefaultResourceLimit) == 0
